// more.go
//
// dta5 extra scripts module
//
// updated 2017-08-18
//
package more

import(
        "dta5/msg"; "dta5/thing"
        "dta5/scripts"
)

// return value of true means stuff should continue
//
// type Script func(obj, subj, dobj, iobj thing.Thing, verb, prep, text string) bool

func Initialize() {
  // "Can't Verb Message, Direct"
  scripts.Scripts["CVMD"] = func(obj, subj, dobj, iobj thing.Thing,
                                verb, prep, text string) bool {
    if obj != dobj { return true }
    
    var mesg string
    var ok bool
    
    if dat := obj.Data("CVMD_" + verb); dat == nil {
      scripts.Log("CVDM(%q, %q): obj.Data() is nil", obj.Ref(), verb)
      return true
    } else {
      if mesg, ok = dat.(string); !ok {
        scripts.Log("CVDM(%q, %q): obj.Data() of wrong type (%T)", obj.Ref(), verb, obj.Ref())
        return true
      }
    }
    
    subj.Deliver(msg.New(mesg))
    return false
  }
}
