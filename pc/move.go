// move.go
//
// dta5 PlayerChar movement actions
//
// updated 2017-08-18
//
package pc

import( "strings";
        "dta5/door"; "dta5/log"; "dta5/msg"; "dta5/name"; "dta5/room"
        "dta5/thing"; "dta5/util";
)

func DoMoveDir(pp *PlayerChar, dir room.NavDir) {
  loc := pp.where.Place.(*room.Room)
  tgt := loc.Nav(dir)
  
  if tgt == nil {
    pp.QWrite("You cannot go %s from here.", cardDirNames[dir])
    return
  }
  
  switch t_tgt := tgt.(type) {
  case *room.Room:
    leave_msg := msg.New("txt", "%s goes %s.", pp.Normal(0), cardDirNames[dir])
    leave_msg.Add(pp, "txt", "You head %s.", cardDirNames[dir])
    loc.Deliver(leave_msg)
    arrive_msg := msg.New("txt", "%s arrives.", pp.Normal(0))
    t_tgt.Deliver(arrive_msg)
    loc.Contents.Remove(pp)
    t_tgt.Contents.Add(pp)
    DoLook(pp, "look", nil, "", nil, "")
    
  case *door.Doorway:
    if t_tgt.IsOpen() {
      sname := util.Cap(pp.Normal(0))
      o_dwy := t_tgt.Other()
      oname := o_dwy.Normal(0)
      var tgt_rm *room.Room
      var ar_m *msg.Message
      
      switch o_cont := o_dwy.Loc().Place.(type) {
      case *room.Room:
        ar_m = msg.New("txt", "%s arrives through %s.", sname, oname)
        tgt_rm = o_cont
      case thing.Container:
        o_cont_t := o_cont.(thing.Thing)
        var no_err bool
        if tgt_rm, no_err = o_cont_t.Loc().Place.(*room.Room); !no_err {
          log(dtalog.ERR, "DoMove(): other *door.Doorway (%q) not contained in a container in a Room.", o_dwy.Ref())
          pp.QWrite("Some unseen force prevents you. (Really, though, this is a game error.)")
          return
        }
        ar_m = msg.New("txt", "%s arrives through %s %s.", sname, o_dwy.Normal(0), o_dwy.Loc().String())
      default:
        log(dtalog.ERR, "DoMove(): other *door.Doorway (%q) not contained in room.Room or in thing.Container in a room.Room.", o_dwy.Ref())
        pp.QWrite("Some unseen force prevents you. (Really, though, this is a game error.)")
        return
      }
      
      lv_m := msg.New("txt", "%s goes %s through %s.", pp.Normal(0),
                            cardDirNames[dir], t_tgt.Normal(0))
      lv_m.Add(pp, "txt", "You head %s through %s.", cardDirNames[dir], t_tgt.Normal(0))
      
      loc.Deliver(lv_m)
      tgt_rm.Deliver(ar_m)
      loc.Contents.Remove(pp)
      tgt_rm.Contents.Add(pp)
      DoLook(pp, "look", nil, "", nil, "")
    } else {
      pp.QWrite("%s is closed.", util.Cap(t_tgt.Normal(name.DEF_ART)))
    }
      
  default:
    pp.QWrite("Sorry, that isn't supported yet.")
  }
}

// type DoFunc func(*PlayerChar,
//                  string,           verb
//                  thing.Thing,      direct object
//                  string,           preposition
//                  thing.Thing,      indirect object
//                  string)           complete command text

func DoMove(pp *PlayerChar, verb string, dobj thing.Thing,
            prep string, iobj thing.Thing, text string) {
  
  loc := pp.where.Place.(*room.Room)
  
  if dobj == nil {
    if iobj == nil {
      pp.QWrite("Go where?")
      return
    }
    
    if prep == "behind" {
      sname := util.Cap(pp.Normal(0))
      oname := iobj.Normal(0)
      m := msg.New("txt", "%s walks behind %s.", sname, oname)
      m.Add(pp, "txt", "You walk behind %s.", oname)
      m.Add(iobj, "txt", "%s walks behind you.", sname)
      loc.Deliver(m)
      return
    } else {
      pp.QWrite("You cannot walk %s %s.", strings.ToUpper(prep), iobj.Normal(0))
      return
    }
  } else {
    if dwy, ok := dobj.(*door.Doorway); !ok {
      sname := util.Cap(pp.Normal(0))
      oname := dobj.Normal(0)
      if iobj == nil {
        m := msg.New("txt", "%s walks toward %s.", sname, oname)
        m.Add(pp, "txt", "You walk toward %s.", oname)
        m.Add(dobj, "txt", "%s walks toward you.", sname)
        loc.Deliver(m)
      } else {
        iname := iobj.Normal(0)
        var m *msg.Message
        switch prep {
        case "in", "on":
          m = msg.New("txt", "%s walks toward %s.", sname, iname)
          m.Add(iobj, "txt", "%s walks toward you.", sname)
        case "behind", "under":
          m = msg.New("txt", "%s walks toward something %s %s.", sname, prep, iname)
          m.Add(iobj, "txt", "%s walks toward something %s you.", sname, prep)
        }
        m.Add(pp, "txt", "You walk toward %s %s %s.", oname, prep, iname)
        loc.Deliver(m)
      }
    } else {
      if dwy.IsOpen() == false {
        pp.QWrite("%s is closed.", util.Cap(dobj.Normal(name.DEF_ART)))
        return
      }
      
      var lv_m, ar_m *msg.Message
      var tgt_rm *room.Room
      var o_dwy = dwy.Other()
      var sname = util.Cap(pp.Normal(0))
      var oname = dwy.Normal(0)
      
      switch o_cont := o_dwy.Loc().Place.(type) {
      case *room.Room:
        tgt_rm = o_cont
        ar_m = msg.New("txt", "%s arrives through %s.", sname, o_dwy.Normal(0))
      case thing.Container:
        o_cont_t := o_cont.(thing.Thing)
        var no_err bool
        if tgt_rm, no_err = o_cont_t.Loc().Place.(*room.Room); !no_err {
          log(dtalog.ERR, "DoMove(): other *door.Doorway (%q) not contained in a container in a Room.", o_dwy.Ref())
          pp.QWrite("Some unseen force prevents you. (Really, though, this is a game error.)")
          return
        }
        ar_m = msg.New("txt", "%s arrives through %s %s.", sname, o_dwy.Normal(0), o_dwy.Loc().String())
      default:
        log(dtalog.ERR, "DoMove(): other *door.Doorway (%q) not contained in room.Room or in thing.Container in a room.Room.", o_dwy.Ref())
        pp.QWrite("Some unseen force prevents you. (Really, though, this is a game error.)")
        return
      }
      
      if iobj == nil {
        lv_m = msg.New("txt", "%s goes through %s.", sname, oname)
        lv_m.Add(pp, "txt", "You go through %s", oname)
      } else {
        prep_loc := dobj.Loc().String()
        lv_m = msg.New("txt", "%s goes through %s %s.", sname, oname, prep_loc)
        lv_m.Add(pp, "txt", "You go through %s %s.", oname, prep_loc)
      }
      
      loc.Deliver(lv_m)
      tgt_rm.Deliver(ar_m)
      loc.Contents.Remove(pp)
      tgt_rm.Contents.Add(pp)
      DoLook(pp, "look", nil, "", nil, "")
    }
  }
}
