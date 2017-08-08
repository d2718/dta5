// move.go
//
// dta5 PlayerChar movement actions
//
// updated 2017-08-05
//
package pc

import(
        "dta5/door"; "dta5/msg"; "dta5/room"
)

var cardDirs map[string]room.NavDir = map[string]room.NavDir {
  "n":  room.N,  "north": room.N,
  "ne": room.NE, "northeast": room.NE,
  "e":  room.E,  "east": room.E,
  "se": room.SE, "southeast": room.SE,
  "s":  room.S,  "south": room.S,
  "sw": room.SW, "southwest": room.SW,
  "w":  room.W,  "west": room.W,
  "nw": room.NW, "northwest": room.NW,
  "u":  room.UP,  "up": room.UP,
  "d":  room.DOWN, "down": room.DOWN,
  "o":  room.OUT, "out": room.OUT, }

var cardDirNames map[room.NavDir]string = map[room.NavDir]string {
  room.N: "north", room.NE: "northeast", room.E: "east",
  room.SE: "southeast", room.S: "south", room.SW: "southwest",
  room.W: "west", room.NW: "northwest", room.UP: "up", room.DOWN: "down",
  room.OUT: "out", }

func DoMoveDir(pp *PlayerChar, dir room.NavDir) {
  loc := pp.where.Place.(*room.Room)
  tgt := loc.Nav(dir)
  
  if tgt == nil {
    pp.QWrite("You cannot go %s from here.", cardDirNames[dir])
    return
  }
  
  switch t_tgt := tgt.(type) {
  case *room.Room:
    leave_msg := msg.New("%s goes %s.", pp.Normal(0), cardDirNames[dir])
    leave_msg.Add(pp, "You head %s.", cardDirNames[dir])
    loc.Deliver(leave_msg)
    loc.Contents.Remove(pp)
    t_tgt.Contents.Add(pp)
    arrive_msg := msg.New("%s arrives.", pp.Normal(0))
    arrive_msg.Add(pp, "")
    t_tgt.Deliver(arrive_msg)
    DoLook(pp, "look", nil, "", nil, "")
    
  case *door.Doorway:
    if t_tgt.IsOpen() {
      other_dwy := t_tgt.Other()
      tgt_loc := other_dwy.Loc().Place.(*room.Room)
      leave_msg := msg.New("%s goes %s through %s.", pp.Normal(0),
                            cardDirNames[dir], t_tgt.Normal(0))
      leave_msg.Add(pp, "You head %s through %s.", cardDirNames[dir], t_tgt.Normal(0))
      loc.Deliver(leave_msg)
      loc.Contents.Remove(pp)
      tgt_loc.Contents.Add(pp)
      arrive_msg := msg.New("%s arrives through %s.", pp.Normal(0), other_dwy.Normal(0))
      arrive_msg.Add(pp, "")
      tgt_loc.Deliver(arrive_msg)
      DoLook(pp, "lool", nil, "", nil, "")
    } else {
      pp.QWrite("%s is closed.", t_tgt.Normal(0))
    }
      
  default:
    pp.QWrite("Sorry, that isn't supported yet.")
  }
}