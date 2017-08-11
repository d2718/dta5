// ref.go
//
// dta5 Referent interface
//
// A Referent is anything that can be universally identified by a string
// reference.
//
// updated 2017-08-11
//
package ref

import( "fmt"; "sync";
        "dta5/log";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("ref: " + fmtstr, args...))
}

type Interface interface {
  Ref() string
}

var referents map[string]Interface
var locker *sync.Mutex

func init() {
  referents = make(map[string]Interface)
  locker = new(sync.Mutex)
}

func Reset() {
  log(dtalog.DBG, "Reset() called")
  referents = make(map[string]Interface)
}

func Register(r Interface) error {
  log(dtalog.DBG, "Register(%q) called", r.Ref())
  locker.Lock()
  defer locker.Unlock()
  if _, exists := referents[r.Ref()]; exists {
    log(dtalog.WRN, "ref.Register(%q): reference already registered; replacing", r.Ref())
  }
  referents[r.Ref()] = r
  return nil
}

func Reregister(r Interface) error {
  log(dtalog.DBG, "Reregister(%q) called", r.Ref())
  locker.Lock()
  defer locker.Unlock()
  if _, exists := referents[r.Ref()]; !exists {
    log(dtalog.WRN, "ref.Reregister(%q): attempting to reregister nonexistent reference; registering", r.Ref())
  }
  referents[r.Ref()] = r
  return nil
}

func Deregister(r Interface) error {
  log(dtalog.DBG, "Deregister(%q) called", r.Ref())
  locker.Lock()
  defer locker.Unlock()
  _, exists := referents[r.Ref()]
  if exists {
    delete(referents, r.Ref())
    return nil
  } else {
    return fmt.Errorf("ref.Deregister([ref.Ref %q]): reference %q not registered",
                      r.Ref(), r.Ref())
  }
}

func Deref(s string) Interface {
  locker.Lock()
  defer locker.Unlock()
  r, exists := referents[s]
  if exists {
    return r
  } else {
    return nil
  }
}

func Walk(f func(Interface)) {
  locker.Lock()
  for _, r := range referents {
    f(r)
  }
  locker.Unlock()
}

// NilGuard() is a debugging function.
//
func NilGuard(r Interface) string {
  if r == nil {
    return "<nil>"
  } else {
    return r.Ref()
  }
}
