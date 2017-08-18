// examine.go
//
// dta5 PlayerChar verb for looking more closely at an object.
//
// updated 2017-08-15
//
package pc

import(
        "github.com/delicb/gstring";
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

func DoExamine(pp *PlayerChar, verb string, dobj thing.Thing,
               prep string, iobj thing.Thing, text string) {
  
  if dobj == nil {
    pp.QWrite("Examine what?")
    return
  }
  
  bod := pp.Body()
  if !bod.IsHolding(dobj) {
    pp.QWrite("You must be holding %s to examine it.", dobj.Normal(name.DEF_ART))
    return
  }
  
  m := msg.New("txt", "%s examines %s %s.", util.Cap(pp.Normal(0)),
                pp.PossPronoun(), dobj.Normal(name.NO_ART))
  m.Add(pp, "txt", "You examine your %s.", dobj.Normal(name.NO_ART))
  pp.where.Place.(*room.Room).Deliver(m)
  
  pp.QWrite("You see %s.", dobj.Full(0))
  pp.QWrite(dobj.Desc())
  
  if t, ok := dobj.(thing.Openable); ok {
    if t.IsToggleable() {
      if t.IsOpen() {
        pp.QWrite("The %s is open.", dobj.Short(name.NO_ART))
      } else {
        pp.QWrite("The %s is closed.", dobj.Short(name.NO_ART))
      }
    }
  }
  
  if t, ok := dobj.(thing.Container); ok {
    store_preps := make([]string, 0, 4)
    for _, s_num := range []byte{thing.IN, thing.ON, thing.BEHIND, thing.UNDER} {
      if t.Side(s_num) != nil {
        store_preps = append(store_preps, thing.SideStr(s_num))
      }
    }
    if len(store_preps) > 0 {
      pp.QWrite("It appears you can store things %s %s.",
                util.EnglishList(store_preps), dobj.Short(name.DEF_ART))
    }
  }
  
  if t, ok := dobj.(thing.Wearable); ok {
    slot := t.Slot()
    num_slots, _ := bod.WornSlots(slot)
    if num_slots > 0 {
      worn_slot_str := bod.WornSlotName(slot)
      if worn_slot_str == "" {
        pp.QWrite("You appear to be able to wear %s.", dobj.Short(name.DEF_ART))
      } else {
        worn_slot := gstring.Sprintm(worn_slot_str, map[string]interface{} { "pp": "your" })
        pp.QWrite("You appear to be able to wear %s %s.",
                  dobj.Short(name.DEF_ART), worn_slot)
      }
    } else {
      pp.QWrite("The %s appears wearable, but not by you.", dobj.Short(name.NO_ART))
    }
  }
}
