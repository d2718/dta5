// desc.go
//
// dta5 managing pages of disk-borne descriptions
//
// updated 2017-08-30
//
// DTA5 deals with a great deal of text; it's the only information that
// players see, and it's used to build the entirety of a fictional world
// in their imaginations. Vast swathes of the game world and the items
// contained therein require paragraphs upon paragraphs of descriptive text.
// Not all of this text has to sit in memory all the time; any given game
// world will undoubtedly be sparsely populated (MUDs just aren't cool
// anymore), so parts that have sat dormant can have their descriptions
// "unloaded" and not reloaded until someone needs to "look" at them again.
//
// The mechanism is thus: Textual descriptions reside in files in a specific
// subdirectory of the game world directory (by default "descs/"). Each file
// contains a "page" of descriptions (yes, this is by analogy to a "page" of
// virtual memory, because the action here is vaguely analogous to swapping).
// Upon initialization, each file gets walked, and each ref.Interface that
// is described has its page pointer set. Thereafter, whenever a thing's
// (by "thing" we mean something implementing desc.Interface) description is
// requested, the package loads the page from disk if necessary, and then
// returns the description.
//
// At each access of a descriptive page in memory, its last access time is
// set. Periodically a sweep is made of all loaded pages, and if a given
// page's last access time is old enough, that page is unloaded from memory.
//
// The format of a description page file is a series of two-element JSON
// lists. The first element is the ref string of the described thingy,
// the second is the description text. For example
//
//  ["r0-t1", "The mailbox is covered in rust-patched, chipped paint."]
//
// Remember: JSON doesn't support multi-line strings, so you will need to use
// the appropriate escape sequence for line breaks.
//
package desc

import( "fmt"; "encoding/json"; "os"; "path/filepath"; "time";
        "dta5/log"; "dta5/ref";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("desc: " + fmtstr, args...))
}

type Interface interface {
  SetDescPage(*string)
  Desc() string
}

// If a page has not had one of its descriptions read after StalePageLife,
// it is officially "stale" and will be unloaded on the next unload check.
//
var StalePageLife time.Duration = time.Duration(time.Minute * time.Duration(5))

// The Page keeps track of how long it has been since its last access, and
// a map of ref strings to their description strings.
//
type Page struct{
  Path string
  Stale time.Time
  Stuff map[string]string
}
 
// The master Page repository.
//
var Pages map[string]*Page = make(map[string]*Page)

// Paths contains all of the paths to files containing pages of descriptions.
// Each desc.Interface locates its descriptive text by storing a pointer to
// the element of this slice that contains its page's path.
//
var Paths []string

// A description file may hold the description of an object that is not
// loaded into memory when the game starts (for example, items in the
// inventory of a player character). When the Initialize() function initially
// walks the description files to correctly set the pointers to the pages
// for each item, the descriptions of not-yet-loaded items are put "in Limbo",
// and when they finally are loaded, they are Unlimbo()'d.
//
var Limbo map[string]*string

// Initialize() iterates through the files in the supplied directory, reading
// each one in turn and setting each desc.Interface-implementing item's
// description pointer to the appropriate page (or putting it in Limbo).
//
func Initialize(basePath string) error {
  log(dtalog.DBG, "Initialize(%q) called", basePath)
  dir, err := os.Open(basePath)
  if err != nil {
    log(dtalog.ERR, "Initialize(%q): error opening directory for read: %s",
                    basePath, err)
    return err
  }
  
  filez, err := dir.Readdirnames(0)
  if err != nil {
    log(dtalog.ERR, "Initialize(%q): error reading from directory: %s",
                    basePath, err)
    dir.Close()
    return err
  }
  dir.Close()
  
  Paths = make([]string, 0, len(filez))
  Limbo = make(map[string]*string)
  
  for _, fname := range filez {
    pth := filepath.Join(basePath, fname)
    f, err := os.Open(pth)
    if err != nil {
      log(dtalog.WRN, "Initialize(): error opening file %q: %s", pth, err)
      continue
    }
    defer f.Close()
    log(dtalog.DBG, "Initialize(): reading file %q", pth)
    
    Paths = append(Paths, pth)
    cur_ptr := &(Paths[len(Paths)-1])
    
    dcdr := json.NewDecoder(f)
    var raw interface{}
    var raw_slice []interface{}
    var i ref.Interface
    var idx string
    for dcdr.More() {
      err = dcdr.Decode(&raw)
      if err != nil {
        log(dtalog.WRN, "Initialize(): error decoding from file %q: %s",
                        pth, err)
        continue
      }
      raw_slice = raw.([]interface{})
      if len(raw_slice) < 2 {
        log(dtalog.WRN, "Initialize(): in file %q: slice too short: %q",
                        pth, raw_slice)
        continue
      }
      
      idx = raw_slice[0].(string)
      i = ref.Deref(idx)
      if i == nil {
        Limbo[idx] = cur_ptr
      } else {
        i.(Interface).SetDescPage(cur_ptr)
      }
    }
  }
  return nil
}

// Sets the description page pointer for a Limbo'd object when it is loaded.
//
func UnLimbo(x Interface) {
  ref_str := x.(ref.Interface).Ref()
  if sptr, isIn := Limbo[ref_str]; isIn {
    x.SetDescPage(sptr)
    delete(Limbo, ref_str)
  }
}

// Load the page whose descriptions are in the given file.
//
func LoadPage(pth string) error {
  log(dtalog.DBG, "LoadPage(%q) called", pth)
  f, err := os.Open(pth)
  if err != nil {
    log(dtalog.ERR, "LoadPath(%q): error opening file: %s", pth, err)
    return err
  }
  defer f.Close()
  
  npagep := &Page{ Path: pth, Stale: time.Now().Add(StalePageLife),
                    Stuff: make(map[string]string), }
  
  dcdr := json.NewDecoder(f)
  
  var raw interface{}
  var raw_slice []interface{}
  for dcdr.More() {
    err = dcdr.Decode(&raw)
    if err != nil {
      log(dtalog.WRN, "LoadPath(%q): decoding error: %s", pth, err)
      continue
    }
    raw_slice = raw.([]interface{})
    npagep.Stuff[raw_slice[0].(string)] = raw_slice[1].(string)
  }
  
  Pages[pth] = npagep
  return nil
}

// Walks through the loaded Pages and deletes the ones that haven't been
// consulted since StalePageLife ago.
//
func UnloadStale() {
  log(dtalog.DBG, "UnloadStale() called")
  stale_keys := make([]string, 0, len(Pages))
  now := time.Now()
  
  for k, pgp := range Pages {
    if pgp.Stale.After(now) {
      stale_keys = append(stale_keys, k)
    }
  }
  
  log(dtalog.DBG, "UnloadStale(): stale keys: %q", stale_keys)
  
  for _, k := range stale_keys {
    delete(Pages, k)
  }
}

// Get the description for the provided Page/thing combo, loading the Page
// from disk if necessary.
//
func GetDesc(pagePath, ref string) string {
  log(dtalog.DBG, "GetDesc(%q, %q) called", pagePath, ref)
  var pgp *Page
  var ok bool
  if pgp, ok = Pages[pagePath]; !ok {
    err := LoadPage(pagePath)
    if err != nil {
      log(dtalog.ERR, "GetDesc(%q, %q): Error in LoadPage(): %s",
                      pagePath, ref, err)
    }
    pgp = Pages[pagePath]
  }
  
  pgp.Stale = time.Now().Add(StalePageLife)
  dscr := pgp.Stuff[ref]
  return dscr
}
