// wear.go
//
// dta5 PlayerChar wear/remove verbs
//
// updated 2017-08-11
//
package pc

import( "fmt";
        "github.com/delicb/gstring";
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

// type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

func DoWear(pp *PlayerChar, verb string, dobj thing.Thing,
            prep string, iobj thing.Thing, text string) {
  
  if dobj == nil {
    pp.QWrite("Wear what?")
    return
  }
  
  bod := pp.Body()
  if !bod.IsHolding(dobj) {
    pp.QWrite("You are not holding %s.", dobj.Normal(name.DEF_ART))
    return
  }
  
  if wt, ok := dobj.(thing.Wearable); ok {
    slot := wt.Slot()
    var slots_worn byte = 0
    var can_wear byte
    can_wear, _ = bod.WornSlots(slot)
    var already_worn = make([]string, 0, 0)
    for _, t := range pp.Inventory.Things {
      if wit, wok := t.(thing.Wearable); wok {
        if !bod.IsHolding(t) {
          if wit.Slot() == slot {
            slots_worn++
            already_worn = append(already_worn, t.Normal(0))
          }
        }
      }
    }
    
    if slots_worn < can_wear {
      if rh, _ := bod.HeldIn("right_hand"); rh == dobj {
        bod.SetHeld("right_hand", nil)
      } else {
        bod.SetHeld("left_hand", nil)
      }
      f1p := map[string]interface{} { "subj": "You",
                                      "verb": "put",
                                      "pp":   "your",
                                      "dobj": dobj.Normal(0), }
      f3p := map[string]interface{} { "subj": util.Cap(pp.Normal(0)),
                                      "verb": "puts",
                                      "pp":   pp.PossPronoun(),
                                      "dobj": f1p["dobj"], }
      templ := fmt.Sprintf("{subj} {verb} {dobj} %s.", bod.WornSlotName(slot))
      
      m := msg.New(gstring.Sprintm(templ, f3p))
      m.Add(pp, gstring.Sprintm(templ, f1p))
      pp.where.Place.(*room.Room).Deliver(m)
    } else {
      f1p := map[string]interface{} { "pp": "your" }
      slot_str := gstring.Sprintm(bod.WornSlotName(slot), f1p)
      pp.QWrite("You are already wearing %s %s.", util.EnglishList(already_worn), slot_str)
    }
  } else {
    pp.QWrite("You cannot wear %s.", dobj.Normal(0))
  }
}

func DoRemove(pp *PlayerChar, verb string, dobj thing.Thing,
              prep string, iobj thing.Thing, text string) {
  
  if dobj == nil {
    pp.QWrite("Remove what?")
    return
  }
  
  bod := pp.Body()
  if bod.IsHolding(dobj) {
    pp.QWrite("You are already holding %s.", dobj.Normal(name.DEF_ART))
    return
  }
  
  if !pp.Inventory.Contains(dobj) {
    pp.QWrite("You are not wearing %s.", dobj.Normal(name.DEF_ART))
    return
  }
  
  rh, _ := bod.HeldIn("right_hand")
  lh, _ := bod.HeldIn("left_hand")
  if (rh != nil) && (lh != nil) {
    pp.QWrite("You don't have a free hand to hold %s.", dobj.Normal(name.DEF_ART))
    return
  }
  
  f1p := map[string]interface{} { "subj": "You",
                                  "pp":   "your",
                                  "verb": "remove",
                                  "dobj": dobj.Normal(0), }
  f3p := map[string]interface{} { "subj": util.Cap(pp.Normal(0)),
                                  "pp":   pp.PossPronoun(),
                                  "verb": "removes",
                                  "dobj": f1p["dobj"], }
  
  var templ string
  if wt, ok := dobj.(thing.Wearable); ok {
    slot := wt.Slot()
    if slot == "misc" {
      templ = "{subj} {verb} {dobj}."
    } else {
      templ = fmt.Sprintf("{subj} {verb} {dobj} from %s.", bod.WornSlotName(slot))
    }
  } else {
    templ = "{subj} {verb} {dobj}"
  }
  
  m := msg.New(gstring.Sprintm(templ, f3p))
  m.Add(pp, gstring.Sprintm(templ, f1p))
  
  if rh == nil {
    bod.SetHeld("right_hand", dobj)
  } else {
    bod.SetHeld("left_hand", dobj)
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}    
