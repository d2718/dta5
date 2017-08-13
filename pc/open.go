// open.go
//
// dta5 PlayerChar open/close verbs
//
// updated 2017-08-13
//
package pc

import(
        "dta5/door"; "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing";
        "dta5/util";
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
        act_msg = msg.New("%s opens %s %s.", util.Cap(pp.Normal(0)),
                          pp.PossPronoun(), dobj.Normal(name.NO_ART))
        act_msg.Add(pp, "You open your %s.", dobj.Normal(name.NO_ART))
      } else {
        act_msg = msg.New("%s open %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
        act_msg.Add(pp, "You open %s.", dobj.Normal(0))
      }
    } else {
      if prep == "on" {
        act_msg = msg.New("%s opens %s on %s.", util.Cap(pp.Normal(0)),
                          dobj.Normal(0), iobj.Normal(0))
        act_msg.Add(pp, "You open %s on %s.", dobj.Normal(0), iobj.Normal(0))
      } else {
        act_msg = msg.New("%s opens something %s %s.", util.Cap(pp.Normal(0)),
                          prep, iobj.Normal(0))
        act_msg.Add(pp, "You open %s %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
      }
    }
    
    loc.Deliver(act_msg)
    t_dobj.SetOpen(true)
    
    if t_dobj, ok := dobj.(*door.Doorway); ok {
      od := t_dobj.Other()
      om := msg.New("%s opens.", util.Cap(od.Normal(0)))
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
        act_msg = msg.New("%s closes %s %s.", util.Cap(pp.Normal(0)),
                          pp.PossPronoun(), dobj.Normal(name.NO_ART))
        act_msg.Add(pp, "You close your %s.", dobj.Normal(name.NO_ART))
      } else {
        act_msg = msg.New("%s closes %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
        act_msg.Add(pp, "You close %s.", dobj.Normal(0))
      }
    } else {
      if prep == "on" {
        act_msg = msg.New("%s closes %s on %s.", util.Cap(pp.Normal(0)),
                          dobj.Normal(0), iobj.Normal(0))
        act_msg.Add(pp, "You close %s on %s.", dobj.Normal(0), iobj.Normal(0))
      } else {
        act_msg = msg.New("%s closes something %s %s.", util.Cap(pp.Normal(0)),
                          prep, iobj.Normal(0))
        act_msg.Add(pp, "You close %s %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
      }
    }
    
    loc.Deliver(act_msg)
    t_dobj.SetOpen(false)
    
    if t_dobj, ok := dobj.(*door.Doorway); ok {
      od := t_dobj.Other()
      om := msg.New("%s closes.", util.Cap(od.Normal(0)))
      od.Loc().Place.(msg.Messageable).Deliver(om)
    }
    
  default:
    pp.QWrite("You cannot close %s.", dobj.Normal(0))
  }
}
