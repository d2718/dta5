// open.go
//
// dta5 PlayerChar open/close verbs
//
// updated 2017-08-18
//
package pc

import(
        "dta5/act"; "dta5/door"; "dta5/msg"; "dta5/name"; "dta5/room";
        "dta5/scripts"; "dta5/thing"; "dta5/util";
)

// type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

func DoOpen(pp *PlayerChar, verb string,
            dobj thing.Thing, prep string, iobj thing.Thing,
            text string) {
  
  if dobj == nil {
    pp.QWrite("Open what now?")
    return
  }
  
  switch t_dobj := dobj.(type) {
  case thing.Openable:
    if t_dobj.IsOpen() {
      pp.QWrite("%s is already open.", util.Cap(dobj.Normal(name.DEF_ART)))
      return
    }
    if t_dobj.IsToggleable() == false {
      pp.QWrite("You cannot open %s.", dobj.Normal(name.DEF_ART))
      return
    }
    
    loc := pp.Loc().Place.(*room.Room)
    var act_msg *msg.Message
    
    if iobj == nil {
      if dobj.Loc().Place == pp {
        act_msg = msg.New("txt", "%s opens %s %s.", util.Cap(pp.Normal(0)),
                          pp.PossPronoun(), dobj.Normal(name.NO_ART))
        act_msg.Add(pp, "txt", "You open your %s.", dobj.Normal(name.NO_ART))
      } else {
        act_msg = msg.New("txt", "%s open %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
        act_msg.Add(pp, "txt", "You open %s.", dobj.Normal(0))
      }
    } else {
      if prep == "on" {
        act_msg = msg.New("txt", "%s opens %s on %s.", util.Cap(pp.Normal(0)),
                          dobj.Normal(0), iobj.Normal(0))
        act_msg.Add(pp, "txt", "You open %s on %s.", dobj.Normal(0), iobj.Normal(0))
      } else {
        act_msg = msg.New("txt", "%s opens something %s %s.", util.Cap(pp.Normal(0)),
                          prep, iobj.Normal(0))
        act_msg.Add(pp, "txt", "You open %s %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
      }
    }
    
    loc.Deliver(act_msg)
    t_dobj.SetOpen(true)
    
    if t_dobj, ok := dobj.(*door.Doorway); ok {
      od := t_dobj.Other()
      om := msg.New("txt", "%s opens.", util.Cap(od.Normal(0)))
      od.Loc().Place.(msg.Messageable).Deliver(om)
    }
    
  default:
    pp.QWrite("You cannot open %s.", dobj.Normal(0))
  }
}

func DoClose(pp *PlayerChar, verb string,
            dobj thing.Thing, prep string, iobj thing.Thing,
            text string) {
  
  if dobj == nil {
    pp.QWrite("Close what now?")
    return
  }
  
  switch t_dobj := dobj.(type) {
  case thing.Openable:
    if t_dobj.IsOpen() == false {
      pp.QWrite("%s is already closed.", util.Cap(dobj.Normal(name.DEF_ART)))
      return
    }
    if t_dobj.IsToggleable() == false {
      pp.QWrite("You cannot close %s.", dobj.Normal(name.DEF_ART))
      return
    }
    
    loc := pp.Loc().Place.(*room.Room)
    var act_msg *msg.Message
    
    if iobj == nil {
      if dobj.Loc().Place == pp {
        act_msg = msg.New("txt", "%s closes %s %s.", util.Cap(pp.Normal(0)),
                          pp.PossPronoun(), dobj.Normal(name.NO_ART))
        act_msg.Add(pp, "txt", "You close your %s.", dobj.Normal(name.NO_ART))
      } else {
        act_msg = msg.New("txt", "%s closes %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
        act_msg.Add(pp, "txt", "You close %s.", dobj.Normal(0))
      }
    } else {
      if prep == "on" {
        act_msg = msg.New("txt", "%s closes %s on %s.", util.Cap(pp.Normal(0)),
                          dobj.Normal(0), iobj.Normal(0))
        act_msg.Add(pp, "txt", "You close %s on %s.", dobj.Normal(0), iobj.Normal(0))
      } else {
        act_msg = msg.New("txt", "%s closes something %s %s.", util.Cap(pp.Normal(0)),
                          prep, iobj.Normal(0))
        act_msg.Add(pp, "txt", "You close %s %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
      }
    }
    
    loc.Deliver(act_msg)
    t_dobj.SetOpen(false)
    
    if t_dobj, ok := dobj.(*door.Doorway); ok {
      od := t_dobj.Other()
      om := msg.New("txt", "%s closes.", util.Cap(od.Normal(0)))
      od.Loc().Place.(msg.Messageable).Deliver(om)
    }
    
  default:
    pp.QWrite("You cannot close %s.", dobj.Normal(0))
  }
}

// This script is to make doors that will automatically swing shut after
// being opened.

func AutoCloseScript(obj, subj, dobj, iobj thing.Thing,
                      verb, prep, text string) bool {
  
  if obj != dobj { return true }
  
  var delay float64
  var ok bool
  var t_dobj thing.Openable
  
  if dat := dobj.Data("auto_close_script_delay"); dat != nil {
    if delay, ok = dat.(float64); !ok {
      scripts.Log("AutoCloseScript(%q, %q, ..., %q): obj.Data() of wrong type (%T)",
                  obj.Ref(), verb, text, dat)
      return true
    }
  } else {
    scripts.Log("AutoCloseScript(%q, %q, ..., %q): obj.Data() is nil",
                obj.Ref(), verb, text, dat)
    return true
  }
  
  if t_dobj, ok = dobj.(thing.Openable); !ok {
    scripts.Log("AutoCloseScript(%q, %q, ..., %q): obj is not thing.Openable",
                obj.Ref(), verb, text)
    return true
  }
  
  var close_func = func() error {
    if t_dobj.IsOpen() {
      t_dobj.SetOpen(false)
      m := msg.New("txt", "%s closes.", util.Cap(dobj.Normal(0)))
      dobj.Loc().Place.(msg.Messageable).Deliver(m)
      
      if dwy, ok := dobj.(*door.Doorway); ok {
        o_dwy := dwy.Other()
        om := msg.New("txt", "%s closes.", util.Cap(o_dwy.Normal(0)))
        o_dwy.Loc().Place.(msg.Messageable).Deliver(om)
      }
    }
    return nil
  }
  
  act.Add(delay, close_func)
  return true
}

func init() {
  scripts.Scripts["auto_close_script"] = AutoCloseScript
}

