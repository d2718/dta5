// misc_verbs.go
//
// misc_verbs.go
//
// dta5 PlayerChar miscellaneous verbs
//
// inventory
// swap
//
// updated 2017-08-10
//
package pc

import( "fmt"; "strings";
        "github.com/delicb/gstring";
        "dta5/door"; "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing";
        "dta5/util";
)

//type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

func DoExits(pp *PlayerChar, verb string, dobj thing.Thing,
              prep string, iobj thing.Thing, text string) {
  loc := pp.where.Place.(*room.Room)
  exit_dirs := make([]string, 0, 0)
  for n, name := range room.NavDirNames {
    switch e := loc.Nav(n).(type) {
    case *room.Room:
      exit_dirs = append(exit_dirs, name)
    case *door.Doorway:
      exit_dirs = append(exit_dirs, fmt.Sprintf("%s (%s)", name, e.Normal(0)))
    }
  }
  
  switch len(exit_dirs) {
  case 0:
    pp.QWrite("There are no obvious exits from here.")
  case 1:
    pp.QWrite("%s is the only obvious exit from here.", util.Cap(exit_dirs[0]))
  default:
    pp.QWrite("Obvious exits from here are: %s.", util.EnglishList(exit_dirs))
  }
}

func DoInventory(pp *PlayerChar, verb string, dobj thing.Thing,
                 prep string, iobj thing.Thing, text string) {
  
  bod := pp.Body()
  rh, _ := bod.HeldIn("right_hand")
  lh, _ := bod.HeldIn("left_hand")
  if rh == nil {
    if lh == nil {
      pp.QWrite("Your hands are empty.")
    } else {
      pp.QWrite("You are holding %s in your left hand.", lh.Full(0))
    }
  } else {
    if lh == nil {
      pp.QWrite("You are holding %s in your right hand.", rh.Full(0))
    } else {
      pp.QWrite("You are holding %s in your right hand and %s in your left hand.",
                rh.Full(0), lh.Full(0))
    }
  }
  
  worn_stuff := make([]string, 0, 0)
  for _, t := range pp.Inventory.Things {
    if wt, ok := t.(thing.Wearable); ok {
      if !bod.IsHolding(t) {
        str := fmt.Sprintf(".   %s %s", t.Normal(0), bod.WornSlotName(wt.Slot()))
        worn_stuff = append(worn_stuff, str)
      }
    }
  }
  if len(worn_stuff) > 0 {
    f1p := map[string]interface{} { "pp": "your" }
    txt := fmt.Sprintf("You are wearing:\n%s", strings.Join(worn_stuff, "\n"))
    pp.QWrite(gstring.Sprintm(txt, f1p))
  } else {
    pp.QWrite("You aren't wearing anything worth mentioning.")
  }
  
}

func DoSwap(pp *PlayerChar, verb string, dobj thing.Thing, prep string,
            iobj thing.Thing, text string) {
  
  bod := pp.Body()
  rh, _ := bod.HeldIn("right_hand")
  lh, _ := bod.HeldIn("left_hand")
  var m *msg.Message = nil
  if rh == nil {
    if lh == nil {
      pp.QWrite("You have nothing to swap.")
      return
    } else {
      m = msg.New("txt", "%s passes %s from %s left to %s right hand.",
                  util.Cap(pp.Normal(0)), lh.Normal(0),
                  pp.PossPronoun(), pp.PossPronoun())
      m.Add(pp, "txt", "You pass %s from your left to your right hand.", lh.Normal(0))
    }
  } else {
    if lh == nil {
      m = msg.New("txt", "%s passes %s from %s right to %s left hand.",
                  util.Cap(pp.Normal(0)), rh.Normal(0),
                  pp.PossPronoun(), pp.PossPronoun())
      m.Add(pp, "txt", "You pass %s from your right to your left hand.", rh.Normal(0))
    } else {
      m = msg.New("txt", "%s swaps %s in %s left hand with %s in %s right hand.",
                  util.Cap(pp.Normal(0)), lh.Normal(name.DEF_ART),
                  pp.PossPronoun(), rh.Normal(name.DEF_ART),
                  pp.PossPronoun())
      m.Add(pp, "txt", "You swap %s in your left hand with %s in your right hand.",
                lh.Normal(name.DEF_ART), rh.Normal(name.DEF_ART))
    }
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
  bod.SetHeld("left_hand", rh)
  bod.SetHeld("right_hand", lh)
}
