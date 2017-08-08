// scripts.go
//
// dta5 scripts
// I don't think this is going to work.
//
package scripts

import( "fmt";
        "dta5/log"; "dta5/save"; "dta5/thing";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("scripts: " + fmtstr, args...))
}

// return value of true means stuff should continue
//
type Script func(obj, subj, dobj, iobj thing.Thing, verb, prep, text string) bool

var Scripts = make(map[string]Script)

// defined in pc/lock.go
//
// "locked_script"
// "lock_unlock_script"

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

func Bind(obj thing.Thing, verb string, ftag string) {
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
