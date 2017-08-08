// desc.go
//
// dta5 managing pages of disk-borne descriptions
//
// updated 2017-07-29
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

var StalePageLife time.Duration = time.Duration(time.Minute * time.Duration(5))

type Page struct{
  Path string
  Stale time.Time
  Stuff map[string]string
}

var Pages map[string]*Page = make(map[string]*Page)
var Paths []string

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
    var i Interface
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
      i = ref.Deref(idx).(Interface)
      i.SetDescPage(cur_ptr)
    }
  }
  return nil
}

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