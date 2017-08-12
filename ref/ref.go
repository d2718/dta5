// ref.go
//
// dta5 Referent interface
//
// A Referent is anything that can be universally identified by a string
// reference. The points of Referents is so that data structures can refer to
// things that don't currently exist in the game world (such as during the
// course of loading). Instead of holding a pointer to the data in question,
// a string is held, which can be "dereferenced" to the appropriate pointer
// when the data in question IS loaded.
//
// Any Referent can also have arbitrary extra data associated with it; each
// string reference can have an associated map[string]interface{} for use
// with nonstandard behavior.
//
// updated 2017-08-12
//
package ref

import( "fmt"; "sync";
        "dta5/log"; "dta5/save";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("ref: " + fmtstr, args...))
}

type Interface interface {
  Ref() string
  Data(string) interface{}
  SetData(string, interface{})
}

// The referents map holds the association between string identifiers and
// pointers to the data referred to by those referents.
//
var referents map[string]Interface

// The Data map holds "arbitrary extra data" associated with Referents. It is
// "public" so it's easier for the dta5/load package to set when loading
// the game's state.
//
var Data map[string]map[string]interface{}

var refLocker, dataLocker *sync.Mutex

func init() {
  referents = make(map[string]Interface)
  Data = make(map[string]map[string]interface{})
  refLocker = new(sync.Mutex)
  dataLocker = new(sync.Mutex)
}

// Reset() should be called before loading of the game's state.
//
func Reset() {
  log(dtalog.DBG, "Reset() called")
  referents = make(map[string]Interface)
  Data = make(map[string]map[string]interface{})
}

// Used to save all the "arbitrary extra data" in a single go.
//
func SaveData(s save.Saver) {
  x := []interface{}{"data", Data, }
  s.Encode(x)
}

// Register(), as the name describes, registers the given referent with its
// reference string. It will emit a warning (but not an error, although I 
// am still considering possibly panicking in this situation) the given
// reference string already points to a referent.
//
func Register(r Interface) error {
  log(dtalog.DBG, "Register(%q) called", r.Ref())
  refLocker.Lock()
  defer refLocker.Unlock()
  if _, exists := referents[r.Ref()]; exists {
    log(dtalog.WRN, "ref.Register(%q): reference already registered; replacing", r.Ref())
  }
  referents[r.Ref()] = r
  return nil
}

// Reregister() is specifically meant to redirect a reference string to point
// at a different referent. This is useful when creating composed objects:
// first create the base object, then reregister its reference string to point
// to the composed object. Will emit a warning if the reference string is
// NOT already in use.
//
func Reregister(r Interface) error {
  log(dtalog.DBG, "Reregister(%q) called", r.Ref())
  refLocker.Lock()
  defer refLocker.Unlock()
  if _, exists := referents[r.Ref()]; !exists {
    log(dtalog.WRN, "ref.Reregister(%q): attempting to reregister nonexistent reference; registering", r.Ref())
  }
  referents[r.Ref()] = r
  return nil
}

// Deregister() removes the association between a reference string and a
// referent, generally because the referent is not longer needed in the
// game world. This, we hope, allows it to eventually be garbage collected.
//
func Deregister(r Interface) error {
  log(dtalog.DBG, "Deregister(%q) called", r.Ref())
  refLocker.Lock()
  defer refLocker.Unlock()
  _, exists := referents[r.Ref()]
  if exists {
    delete(referents, r.Ref())
    return nil
  } else {
    return fmt.Errorf("ref.Deregister([ref.Ref %q]): reference %q not registered",
                      r.Ref(), r.Ref())
  }
}

// Deref() returns the referent (generally a pointer to some object) associated
// with the given reference string. This is by analogy to "dereferencing" a
// pointer. (I appreciate the irony that this function generally, in fact,
// RETURNS a pointer.)
//
func Deref(s string) Interface {
  refLocker.Lock()
  defer refLocker.Unlock()
  r, exists := referents[s]
  if exists {
    return r
  } else {
    return nil
  }
}

// Walk() calls the supplied function on every Register()'d referent.
//
func Walk(f func(Interface)) {
  refLocker.Lock()
  for _, r := range referents {
    f(r)
  }
  refLocker.Unlock()
}

// GetData() is the helper function called by Interface.Data()
//
func GetData(r Interface, key string) interface{} {
  dataLocker.Lock()
  defer dataLocker.Unlock()
  submap := Data[r.Ref()]
  if submap == nil {
    return nil
  } else {
    return submap[key]
  }
}

// SetData() is the helper function called by Interface.SetData()
//
func SetData(r Interface, key string, val interface{}) {
  r_str := r.Ref()
  dataLocker.Lock()
  defer dataLocker.Unlock()
  if _, ok := Data[r_str]; !ok {
    Data[r_str] = make(map[string]interface{})
  }
  Data[r_str][key] = val
}

// NilGuard() is a debugging function; its purpose should be obvious.
//
func NilGuard(r Interface) string {
  if r == nil {
    return "<nil>"
  } else {
    return r.Ref()
  }
}
