// get.go
//
// dta5 PlayerChar get, put verbs
//
// updated 2017-08-18
//
package pc

import(
        "dta5/log";
        "dta5/msg"; "dta5/name"; "dta5/ref"; "dta5/room";
        "dta5/thing"; "dta5/util";
)

func DoGet(pp *PlayerChar, verb string,
          dobj thing.Thing, prep string, iobj thing.Thing,
          text string) {
  
  log(dtalog.DBG, "DoGet(%q, %q, %q, %q, %q, %q) called",
      ref.NilGuard(pp), verb, ref.NilGuard(dobj), prep,
      ref.NilGuard(iobj), text)
  
  bod := pp.Body()
  rh, _ := bod.HeldIn("right_hand")
  lh, _ := bod.HeldIn("left_hand")
  if (rh != nil) && (lh != nil) {
    pp.QWrite("You don't have a free hand to pick anything up.")
    return
  }
  
  if dobj == nil {
    if iobj == nil {
      pp.QWrite("Get what now?")
    } else {
      pp.QWrite("Get what from %s %s now?", prep, iobj.Normal(name.DEF_ART))
    }
    return
  }
  
  if (dobj.Mass().VT != thing.VT_LTD) || (dobj.Bulk().VT != thing.VT_LTD) {
    pp.QWrite("You cannot pick up %s.", dobj.Normal(name.DEF_ART))
    return
  }
  
  var revealed = make([]string, 0, 0)
  
  if iobj == nil {
    loc := dobj.Loc()
    r := loc.Place.(*room.Room)
    if loc.Side == room.SCENERY {
      r.Scenery.Remove(dobj)
    } else {
      r.Contents.Remove(dobj)
    }
    
    if cont, ok := dobj.(thing.Container); ok {
      for _, s := range []byte{thing.BEHIND, thing.UNDER} {
        tlo := cont.Side(s)
        if tlo != nil {
          temp_t := make([]thing.Thing, 0, len(tlo.Things))
          temp_t = append(temp_t, tlo.Things...)
          for _, t := range temp_t {
            tlo.Remove(t)
            r.Contents.Add(t)
            revealed = append(revealed, t.Normal(0))
          }
        }
      }
    }
    
    mesg := msg.New("txt", "%s picks up %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
    mesg.Add(pp, "txt", "You pick up %s.", dobj.Normal(0))
    r.Deliver(mesg)
    
    if len(revealed) > 0 {
      mesg = msg.New("txt", "Picking up %s reveals %s.", dobj.Short(name.DEF_ART),
                      util.EnglishList(revealed))
      r.Deliver(mesg)
    }
    
  } else {
    cont := iobj.(thing.Container)
    tl := cont.Side(parsePreps[prep])
    tl.Remove(dobj)
    
    if cont, ok := dobj.(thing.Container); ok {
      for _, s := range []byte{thing.BEHIND, thing.UNDER} {
        tlo := cont.Side(s)
        if tlo != nil {
          temp_t := make([]thing.Thing, 0, len(tlo.Things))
          temp_t = append(temp_t, tlo.Things...)
          for _, t := range temp_t {
            tlo.Remove(t)
            tl.Add(t)
            revealed = append(revealed, t.Normal(0))
          }
        }
      }
    }
    
    mesg := msg.New("txt", "%s gets a %s from %s %s.", util.Cap(pp.Normal(0)),
                    dobj.Normal(0), prep, iobj.Normal(0))
    mesg.Add(pp, "txt", "You get a %s from %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
    pp.where.Place.(*room.Room).Deliver(mesg)
    
    if len(revealed) > 0 {
      pp.QWrite("Getting %s reveals %s.", dobj.Short(name.DEF_ART),
                util.EnglishList(revealed))
    }
  }
  
  if rh, _ := bod.HeldIn("right_hand"); rh == nil {
    bod.SetHeld("right_hand", dobj)
  } else {
    bod.SetHeld("left_hand", dobj)
  }
  
  pp.Inventory.Add(dobj)
}


func DoPut(pp *PlayerChar, verb string,
           dobj thing.Thing, prep string, iobj thing.Thing,
           text string) {
  
  bod := pp.Body()
  if dobj == nil {
    if rh, _ := bod.HeldIn("right_hand"); rh != nil {
      dobj = rh
    } else if lh, _ := bod.HeldIn("left_hand"); lh != nil {
      dobj = lh
    } else {
      pp.QWrite("You are not holding anything!")
      return
    }
  }
  
  rh, _ := bod.HeldIn("right_hand")
  lh, _ := bod.HeldIn("left_hand")
  if (dobj != rh) && (dobj != lh) {
    pp.QWrite("You are not holding %s.", dobj.Normal(0))
    return
  }
  
  if iobj != nil {
    switch t_iobj := iobj.(type) {
    case thing.Container:
      prep_side := parsePreps[prep]
      sid := t_iobj.Side(prep_side)
      if sid == nil {
        pp.QWrite("You cannot put anything %s %s.", prep, iobj.Normal(name.DEF_ART))
        return
      }
      if !sid.WillFitBulk(dobj) {
        pp.QWrite("%s is too bulky to fit %s %s.", util.Cap(dobj.Normal(name.DEF_ART)),
                  prep, iobj.Normal(name.DEF_ART))
        return
      }
      if !sid.WillFitMass(dobj) {
        pp.QWrite("%s is too heavy to fit %s %s.", util.Cap(dobj.Normal(name.DEF_ART)),
                  prep, iobj.Normal(name.DEF_ART))
        return
      }
      
      if bod.IsHolding(iobj) {
        if (prep == "behind") || (prep == "under") {
          pp.QWrite("You'll have to put %s down first.", iobj.Normal(name.DEF_ART))
          return
        }
      }
      
      if rh, _ := bod.HeldIn("right_hand"); dobj == rh {
        bod.SetHeld("right_hand", nil)
      } else {
        bod.SetHeld("left_hand", nil)
      }
      pp.Inventory.Remove(dobj)
      sid.Add(dobj)
      
      put_msg := msg.New("txt", "%s puts %s %s %s.", util.Cap(pp.Normal(0)),
                          dobj.Normal(0), prep, iobj.Normal(0))
      put_msg.Add(pp, "txt", "You put %s %s %s.", dobj.Normal(0), prep, iobj.Normal(0))
      pp.where.Place.(*room.Room).Deliver(put_msg)
      return
    default:
      pp.QWrite("You cannot put anything %s %s.", prep, iobj.Normal(0))
      return
    }
  } else {
    rm := pp.where.Place.(*room.Room)
  
    if rh, _ := bod.HeldIn("right_hand"); rh == dobj {
      bod.SetHeld("right_hand", nil)
    } else {
      bod.SetHeld("left_hand", nil)
    }
    pp.Inventory.Remove(dobj)
    rm.Contents.Add(dobj)
    
    put_msg := msg.New("txt", "%s drops %s.", util.Cap(pp.Normal(0)), dobj.Normal(0))
    put_msg.Add(pp, "txt", "You drop %s.", dobj.Normal(0))
    rm.Deliver(put_msg)
    return
  }
}
