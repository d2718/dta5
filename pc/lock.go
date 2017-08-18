// lock.go
//
// dta5 PlayerChar locking and unlocking things.
//
// updated 2017-08-18
//
package pc

import(
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/scripts";
        "dta5/thing"; "dta5/util";
)

func DoLock(pp *PlayerChar, verb string,
            dobj thing.Thing, prep string, iobj thing.Thing,
            text string) {
  
  if dobj == nil {
    if iobj == nil {
      pp.QWrite("%s what?", util.Cap(verb))
    } else if prep == "with" {
      if pp.InHand(iobj) {
        pp.QWrite("What did you want to %s with %s?", verb, iobj.Normal(name.DEF_ART))
      } else {
        pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
      }
    } else {
      pp.QWrite("You can't %s something \"%s\" %s.", verb, prep, iobj.Normal(0))
    }
  } else {
    if iobj == nil {
      pp.QWrite("With what did you want to %s %s?", verb, dobj.Normal(name.DEF_ART))
    } else if prep == "with" {
      if pp.InHand(iobj) {
        pp.QWrite("You cannot %s %s with your %s.", verb,
                  dobj.Normal(name.DEF_ART), iobj.Normal(name.NO_ART))
      } else {
        pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
      }
    } else {
      pp.QWrite("How can you %s %s \"%s\" %s?", verb,
                dobj.Normal(0), prep, iobj.Normal(0))
    }
  }
}

func LockedScript(obj, subj, dobj, iobj thing.Thing,
                  verb, prep, text string) bool {
  if obj != dobj { return true }
  
  if dat := dobj.Data("locked_script_unlocked"); dat != nil {
    if t_dat, ok := dat.(bool); ok {
      if t_dat == true {
        return true
      }
    } else {
      scripts.Log("LockedScript(%q, %q, ... %q): obj.Data() is of wrong type (%T)",
                  obj.Ref(), verb, text, dat)
    }
  } else {
    scripts.Log("LockedScript(%q, %q, ... %q): obj.Data() is nil",
                obj.Ref(), verb, text)
  }
  
  m := msg.New("txt", "%s appears to be locked.", util.Cap(dobj.Normal(name.DEF_ART)))
  subj.Deliver(m)
  return false
}

func LockUnlockScript(obj, subj, dobj, iobj thing.Thing,
                      verb, prep, text string) bool {
                        
  if obj != dobj { return true }
  if iobj == nil { return true }
  
  var pp *PlayerChar
  var ok bool
  
  if pp, ok = subj.(*PlayerChar); !ok { return true }
  
  if !pp.InHand(iobj) { return true }
  
  if dat := dobj.Data("lock_unlock_script_key"); dat != nil {
    switch t_dat := dat.(type) {
    case string:
      if t_dat == iobj.Ref() {
        if dat := dobj.Data("locked_script_unlocked"); dat != nil {
          switch unlkd := dat.(type) {
          case bool:
            if unlkd {
              if verb == "unlock" {
                m := msg.New("%s is already unlocked.",
                              util.Cap(dobj.Normal(name.DEF_ART)))
                subj.Deliver(m)
              } else if verb == "lock" {
                if t_dobj, is_openable := dobj.(thing.Openable); is_openable {
                  if t_dobj.IsOpen() {
                    m := msg.New("txt,", "%s is currently open.",
                                  util.Cap(dobj.Normal(name.DEF_ART)))
                    subj.Deliver(m)
                    return false
                  }
                }
                dobj.SetData("locked_script_unlocked", false)
                m := msg.New("txt", "%s locks %s with %s %s.", util.Cap(subj.Normal(0)),
                              dobj.Normal(0), subj.PossPronoun(),
                              iobj.Normal(name.NO_ART))
                m.Add(pp, "txt", "You lock %s with your %s.", dobj.Normal(0),
                            iobj.Normal(name.NO_ART))
                pp.where.Place.(*room.Room).Deliver(m)
              }
            } else {
              if verb == "unlock" {
                dobj.SetData("locked_script_unlocked", true)
                m := msg.New("txt", "%s unlocks %s with %s %s.", util.Cap(subj.Normal(0)),
                            dobj.Normal(0), subj.PossPronoun(),
                            iobj.Normal(name.NO_ART))
                m.Add(pp, "txt", "You unlock %s with your %s.", dobj.Normal(0),
                            iobj.Normal(name.NO_ART))
                pp.where.Place.(*room.Room).Deliver(m)
              } else {
                m := msg.New("txt", "%s is already locked.",
                              util.Cap(dobj.Normal(name.DEF_ART)))
                subj.Deliver(m)
              }
            }
          default:
            scripts.Log("LockUnlockScript(%q): obj.Data(\"locked_script_unlocked\") is wrong type (%T)", obj.Ref(), dat)
            return true
          }
        } else {
          scripts.Log("LockUnlockScript(%q): obj.Data(\"locked_script_unlocked\") is nil)", obj.Ref(), dat)
          return true
        }
        return false
      } else {
        return true
      }
    default:
      scripts.Log("LockUnlockScript(%q): obj.Data(\"lock_unlock_script_key\") is wrong type (%T)", obj.Ref(), dat)
    }
  } else {
    scripts.Log("LockUnlockScript(%q): obj.Data() is nil", obj.Ref())
  }
  
  return true
}

func init() {
  scripts.Scripts["locked_script"] = LockedScript
  scripts.Scripts["lock_unlock_script"] = LockUnlockScript
}
