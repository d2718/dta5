// build.go
//
// extra dta5 world-building shortcuts
//
// updated 2017-08-18
//
package build

import( "fmt";
        "dta5/log"; "dta5/ref"; "dta5/scripts"; "dta5/thing";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("load/build: " + fmtstr, args...))
}

type Func func([]interface{}) error

var Funx = map[string]Func {
  "key":        MakeKeyAndLocker,
  "autoclose":  MakeAutoClosing,
  "cvmd":       AddCannotDirectMessage,
}

func Build(data []interface{}) error {
  k := data[0].(string)
  if f, ok := Funx[k]; ok {
    return f(data[1:])
  } else {
    log(dtalog.WRN, "Build(%q): unknown function key %q", k)
    return fmt.Errorf("unknown build function %q", k)
  }
}
  

// MakeKeyAndLocker()
//
// ["key_ref", "locker_ref", is_initially_locked ]
//
func MakeKeyAndLocker(data []interface{}) error {
  keyRef   := data[0].(string)
  lockRef  := data[1].(string)
  isLocked := data[2].(bool)
  
  lock := ref.Deref(lockRef).(thing.Thing)
  
  lock.SetData("locked_script_unlocked", !isLocked)
  lock.SetData("lock_unlock_script_key", keyRef)
  scripts.Bind(lock, "open", "locked_script")
  scripts.Bind(lock, "lock", "lock_unlock_script")
  scripts.Bind(lock, "unlock", "lock_unlock_script")
  return nil
}

// MakeAutoClosing()
//
// ["key_ref", delay_secs]
//
func MakeAutoClosing(data []interface{}) error {
  keyRef := data[0].(string)
  delay  := data[1].(float64)
  
  cont := ref.Deref(keyRef)
  cont.SetData("auto_close_script_delay", delay)
  scripts.Bind(cont, "open", "auto_close_script")
  return nil
}

// AddCannotDirectMessage()
//
// ["key_ref", "verb", "message"]

func AddCannotDirectMessage(data []interface{}) error {
  key_ref := data[0].(string)
  verb    := data[1].(string)
  mesg    := data[2].(string)
  
  dat_str := "CVMD_" + verb
  obj     := ref.Deref(key_ref)
  obj.SetData(dat_str, mesg)
  scripts.Bind(obj, verb, "CVMD")
  return nil
}
