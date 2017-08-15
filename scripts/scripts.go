// scripts.go
//
// dta5 scripts
// I don't think this is going to work.
//
package scripts

import( "fmt";
        "dta5/log"; "dta5/ref"; "dta5/save"; "dta5/thing";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("scripts: " + fmtstr, args...))
}

func Log(fmtstr string, args ...interface{}) {
  dtalog.Log(dtalog.WRN, fmt.Sprintf("_script_error_: " + fmtstr, args...))
}

// return value of true means stuff should continue
//
type Script func(obj, subj, dobj, iobj thing.Thing, verb, prep, text string) bool

var Scripts = make(map[string]Script)
//
// defined in pc/get.go
//
//   * "CGMS" ("Cannot Get Messsage Script")
//     Provides custom messaging informing a player why a thing.Thing can't
//     be picked up.
//
// defined in pc/lock.go
//
//   * "locked_script"
//     Prevents a player from opening a thing.Openable and informs him that
//     it's locked, unless the associated data value "locked_script_unlocked"
//     is set to true.
//
//   * "lock_unlock_script"
//     Designates a thing.Thing to be a "key" that allows a thing.Openable
//     under the influence of the "locked_script" to be UNLOCKed by it.
//
// defined in pc/open.go
//
//   * "auto_close_script"
//     Causes a thing.Openable to shut automatically after it has been opened
//    (after a specified delay).
//

var Bindings = make(map[string]map[string]string)

func Check(subj, dobj, iobj thing.Thing, verb, prep, text string) bool {
  
  if iobj != nil {
    if f, sok := Scripts[Bindings[iobj.Ref()][verb]]; sok {
      cont := f(iobj, subj, dobj, iobj, verb, prep, text)
      if !cont {
        return false
      }
    }
  }
  
  if dobj != nil {
    if f, sok := Scripts[Bindings[dobj.Ref()][verb]]; sok {
      cont := f(dobj, subj, dobj, iobj, verb, prep, text)
      if !cont {
        return false
      }
    }
  }
  
  return true
}

func Bind(obj ref.Interface, verb string, ftag string) {
  if _, ok := Scripts[ftag]; ok {
    if _, mok := Bindings[obj.Ref()]; !mok {
      Bindings[obj.Ref()] = make(map[string]string)
    }
    Bindings[obj.Ref()][verb] = ftag
  } else {
    log(dtalog.WRN, "Bind(%q, %q, %q): %q is not a script tag", obj.Ref(), verb, ftag)
  }
}

func SaveBindings(s save.Saver) {
  for t_ref, vmap := range Bindings {
    for verb, script_tag := range vmap {
      data := []interface{}{"script", t_ref, verb, script_tag}
      s.Encode(data)
    }
  }
}
