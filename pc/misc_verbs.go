// misc_verbs.go
//
// misc_verbs.go
//
// dta5 PlayerChar miscellaneous verbs
//
// inventory
// swap
//
// updated 2017-08-05
//
package pc

import(
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

//type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

func DoExits(pp *PlayerChar, verb string, dobj thing.Thing,
              prep string, iobj thing.Thing, text string) {
  loc := pp.where.Place.(*room.Room)
  exit_dirs := make([]string, 0, 0)
  for _, dir := range loc.ExitDirs() {
    exit_dirs = append(exit_dirs, room.NavDirNames[dir])
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
                   
  if pp.RightHand == nil {
    if pp.LeftHand == nil {
      pp.QWrite("You have nothing.")
    } else {
      pp.QWrite("You are holding %s in your left hand.",
                pp.LeftHand.Full(0))
    }
  } else {
    if pp.LeftHand == nil {
      pp.QWrite("You are holding %s in your right hand.",
                pp.RightHand.Full(0))
    } else {
      pp.QWrite("You are holding %s in your right hand and %s in your left hand.",
                pp.RightHand.Full(0), pp.LeftHand.Full(0))
    }
  }
}

func DoSwap(pp *PlayerChar, verb string, dobj thing.Thing, prep string,
            iobj thing.Thing, text string) {
              
  var m *msg.Message = nil
  if pp.RightHand == nil {
    if pp.LeftHand == nil {
      pp.QWrite("You have nothing to swap.")
      return
    } else {
      m = msg.New("%s passes %s from %s left to %s right hand.",
                  util.Cap(pp.Normal(0)), pp.LeftHand.Normal(0),
                  pp.PossPronoun(), pp.PossPronoun())
      m.Add(pp, "You pass %s from your left to your right hand.",
                pp.LeftHand.Normal(0))
    }
  } else {
    if pp.LeftHand == nil {
      m = msg.New("%s passes %s from %s right to %s left hand.",
                  util.Cap(pp.Normal(0)), pp.RightHand.Normal(0),
                  pp.PossPronoun(), pp.PossPronoun())
      m.Add(pp, "You pass %s from your right to your left hand.",
                pp.RightHand.Normal(0))
    } else {
      m = msg.New("%s swaps %s in %s left hand with %s in %s right hand.",
                  util.Cap(pp.Normal(0)), pp.LeftHand.Normal(name.DEF_ART),
                  pp.PossPronoun(), pp.RightHand.Normal(name.DEF_ART),
                  pp.PossPronoun())
      m.Add(pp, "You swap %s in your left hand with %s in your right hand.",
                pp.LeftHand.Normal(name.DEF_ART), pp.RightHand.Normal(name.DEF_ART))
    }
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
  pp.LeftHand, pp.RightHand = pp.RightHand, pp.LeftHand
}