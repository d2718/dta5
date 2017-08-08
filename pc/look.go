// look.go
//
// dta5 PlayerChar look verb
//
// updated 2017-08-06
//
package pc

import( "fmt";
        "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
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
      switch t_dobj := dobj.(type) {
      case *PlayerChar:
        if t_dobj.RightHand != nil {
          if t_dobj.LeftHand != nil {
            pp.QWrite("%s is holding %s in %s right hand and %s in %s left hand.",
                      util.Cap(t_dobj.SubjPronoun()), t_dobj.RightHand.Full(0),
                      t_dobj.PossPronoun(), t_dobj.LeftHand.Full(0),
                      t_dobj.PossPronoun())
          } else {
            pp.QWrite("%s is holding %s in %s right hand.",
                      util.Cap(t_dobj.SubjPronoun()), t_dobj.RightHand.Full(0),
                      t_dobj.PossPronoun())
          }
        } else {
          if t_dobj.LeftHand != nil {
            pp.QWrite("%s is holding %s in %s left hand.",
                      util.Cap(t_dobj.SubjPronoun()), t_dobj.LeftHand.Full(0),
                      t_dobj.PossPronoun())
          }
        }
      case thing.Container:
        s := t_dobj.Side(thing.ON)
        if s != nil {
          if len(s.Things) > 0 {
            pp.QWrite("On %s you see %s.", dobj.Short(name.DEF_ART), s.EnglishList())
          }
        }
      }
      switch t_dobj := dobj.(type) {
      case thing.Openable:
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
