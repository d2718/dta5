// more.go
//
// dta5 extra scripts module
//
// updated 2017-08-29
//
// This is the default package for defining custom behavior. Any function of
// the appropriate signature (see dta5/scripts.Script) can be used to provide
// custom behavior; for sanity's sake they have been collected here by default.
// Feel free to add your own scripts to this package or to add your own
// package.
//
package more

import(
        "dta5/msg"; "dta5/thing"
        "dta5/scripts"
)

// type script.Script func(obj, subj, dobj, iobj thing.Thing, 
//                         verb, prep, text string) bool
//
// Return value of true means the action should continue as normal.
//
// See the documentation for dta5/scripts for explanations of parameters.

// Initialize() should be called on game start (before the world is loaded)
// to associate these scripts with their identifying strings in the scripts
// module. Functions can either be defined directly here, or defined outside
// of this function and added to the scripts.Scripts map[string]Script here.
//
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
    
    subj.Deliver(msg.New("txt", mesg))
    return false
  }
}
