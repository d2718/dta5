// look.go
//
// dta5 PlayerChar look verb
//
// updated 2017-08-11
//
package pc

import( "fmt"; "strings";
        "github.com/delicb/gstring";
        "dta5/body"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

func DoLook(pp *PlayerChar, verb string,
            dobj thing.Thing, prep string, iobj thing.Thing,
            text string) {
  
  if iobj != nil {
    if dobj != nil {
      pp.QWrite("You look at %s %s %s.", dobj.Full(0), prep, iobj.Normal(0))
      pp.QWrite(dobj.Desc())
      switch t_dobj := dobj.(type) {
      case thing.Openable:
        if t_dobj.IsOpen() {
          if t_dobj.IsOpen() {
            pp.QWrite("%s is open.", util.Cap(dobj.Short(name.DEF_ART)))
          } else {
            pp.QWrite("%s is closed.", util.Cap(dobj.Short(name.DEF_ART)))
          }
        }
      default:
        // do nothing else special
      }
    } else {
      t_iobj := iobj.(thing.Container)
      stuff := t_iobj.Side(parsePreps[prep]).EnglishList()
      pp.QWrite("%s %s you see %s.", util.Cap(prep), t_iobj.Normal(name.DEF_ART), stuff)
    }
  } else {
    if dobj != nil {
      pp.QWrite("You look at %s.", dobj.Full(0))
      pp.QWrite(dobj.Desc())

      if t_dobj, ok := dobj.(body.Bodied); ok {
        b := t_dobj.Body()
        fmap := map[string]interface{} { "pp": dobj.PossPronoun(), }
        held_stuff := make([]string, 0, 0)
        for _, slot := range b.HeldSlotKeys() {
          t, _ := b.HeldIn(slot)
          if t != nil {
            fmap["obj"] = t.Normal(0)
            str := gstring.Sprintm("{obj} " + b.HeldSlotName(slot), fmap)
            held_stuff = append(held_stuff, str)
          }
        }
        
        if len(held_stuff) > 0 {
          held_stuff_list := util.EnglishList(held_stuff)
          subj_p := util.Cap(dobj.SubjPronoun())
          pp.QWrite("%s is holding %s.", subj_p, held_stuff_list)
        }
      }
      
      if t_dobj, ok := dobj.(*PlayerChar); ok {
        worn_stuff := make([]string, 0, 0)
        b := t_dobj.Body()
        for _, t := range t_dobj.Inventory.Things {
          if wt, ok := t.(thing.Wearable); ok {
            if b.IsHolding(t) == false {
              slot := wt.Slot()
              worn_stuff = append(worn_stuff, fmt.Sprintf(".   %s %s", t.Normal(0), b.WornSlotName(slot)))
            }
          }
        }
        if len(worn_stuff) > 0 {
          fmap := map[string]interface{} {
            "pp": dobj.PossPronoun(),
            "sp": util.Cap(dobj.SubjPronoun()),
          }
          txt := fmt.Sprintf("{sp} is wearing: \n%s", strings.Join(worn_stuff, "\n"))
          pp.QWrite(gstring.Sprintm(txt, fmap))
        }
      }
        
      if t_dobj, ok := dobj.(thing.Container); ok {
        s := t_dobj.Side(thing.ON)
        if s != nil {
          if len(s.Things) > 0 {
            pp.QWrite("On %s you see %s.", dobj.Short(name.DEF_ART), s.EnglishList())
          }
        }
      }
      
      if t_dobj, ok := dobj.(thing.Openable); ok {
        if t_dobj.IsToggleable() {
          if t_dobj.IsOpen() {
            pp.QWrite("%s is open.", util.Cap(dobj.Short(name.DEF_ART)))
          } else {
            pp.QWrite("%s is closed.", util.Cap(dobj.Short(name.DEF_ART)))
          }
        }
      }
      
    } else {
              
      loc := pp.where.Place.(*room.Room)
      
      rm_name := loc.Title
      rm_text := loc.Desc()
      
      t_tl := pp.AllButMe()
      
      var also string
      if len(t_tl.Things) > 0 {
        also = fmt.Sprintf("\n\nYou also see %s.", t_tl.EnglishList())
      } else {
        also = ""
      }
      
      pp.Send(PM{Type: "headline", Payload: rm_name})
      pp.QWrite("\n* %s *\n\n%s%s", rm_name, rm_text, also)
    }
  }
}
