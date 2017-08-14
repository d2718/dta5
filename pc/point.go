// point.go
//
// dta5 PlayerChar pointing, waving, etc.
//
// updated 2017-08-14
//
package pc

import(
        "github.com/delicb/gstring";
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

// see pc/emote.go for
//   * struct verbConj
//   * map[room.NavDir]string dirDirMsgs

var pointVerbFilter = map[string]verbConj {
}

var pointDirDefaultTemplate = "{subj} {verb} {dir}."
var pointDirWithTemplate = "{subj} {verb} {subj_pp} {iobj} {dir}."

func DoPointDir(pp *PlayerChar, verb string, dirNum room.NavDir,
                iobj thing.Thing) {
                  
  var f1p = map[string]interface{} { "subj": "You", "subj_pp": "your", }
  var f3p = map[string]interface{} { "subj": util.Cap(pp.Normal(0)),
                                     "subj_pp":   pp.PossPronoun(), }
  if v, ok := pointVerbFilter[verb]; ok {
    f1p["verb"], f3p["verb"] = v.p1, v.p3
  } else {
    f1p["verb"], f3p["verb"] = verb, verb + "s"
  }
  var dir1p = gstring.Sprintm(dirDirMsgs[dirNum], f1p)
  var dir3p = gstring.Sprintm(dirDirMsgs[dirNum], f3p)
  f1p["dir"], f3p["dir"] = dir1p, dir3p
  
  var m *msg.Message
  
  if iobj == nil {
    m = msg.New(gstring.Sprintm(pointDirDefaultTemplate, f3p))
    m.Add(pp, gstring.Sprintm(pointDirDefaultTemplate, f1p))
  } else {
    if !pp.Body().IsHolding(iobj) {
      pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
      return
    }
    
    f1p["iobj"] = iobj.Normal(name.NO_ART)
    f3p["iobj"] = f1p["iobj"]
    
    m = msg.New(gstring.Sprintm(pointDirWithTemplate, f3p))
    m.Add(pp, gstring.Sprintm(pointDirWithTemplate, f1p))
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}

var pointIntransTemplate     = "{subj} {verb}."
var pointIntransWithTemplate = "{subj} {verb} with {subj_pp} {iobj}."
var pointIndOnlyTemplate     = "{subj} {verb} at the space {prep} {iobj}."
var pointTransTemplate       = "{subj} {verb} at {dobj}."
var pointTransWithTemplate   = "{subj} {verb} {subj_pp} {iobj} at {dobj}."

func DoPoint(pp *PlayerChar, verb string, dobj thing.Thing,
             prep string, iobj thing.Thing, text string) {
  
  var f1p = map[string]interface{} { "subj": "You", "subj_pp": "your", }
  var f3p = map[string]interface{} { "subj": util.Cap(pp.Normal(0)),
                                     "subj_pp": pp.PossPronoun(), }
  if v, ok := pointVerbFilter[verb]; ok {
    f1p["verb"], f3p["verb"] = v.p1, v.p3
  } else {
    f1p["verb"], f3p["verb"] = verb, verb + "s"
  }
  
  var m *msg.Message
  
  if dobj == nil {
    
    if iobj == nil {
      m = msg.New(gstring.Sprintm(pointIntransTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointIntransTemplate, f1p))
      
    } else if prep == "with" {
      if !pp.Body().IsHolding(iobj) {
        pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
        return
      }
      f1p["iobj"] = iobj.Normal(name.NO_ART)
      f3p["iobj"] = f1p["iobj"]
      m = msg.New(gstring.Sprintm(pointIntransWithTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointIntransWithTemplate, f1p))
      
    } else {
      f1p["iobj"] = iobj.Normal(0)
      f3p["iobj"] = f1p["iobj"]
      f1p["prep"], f3p["prep"] = prep, prep
      m = msg.New(gstring.Sprintm(pointIndOnlyTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointIndOnlyTemplate, f1p))
    }
    
  } else if dobj == pp {
    f1p["dobj"] = "yourself"
    f3p["dobj"] = pp.ReflexPronoun()
    
    if iobj == nil {
      m = msg.New(gstring.Sprintm(pointTransTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointTransTemplate, f1p))
    } else if prep == "with" {
      if !pp.Body().IsHolding(iobj) {
        pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
        return
      }
      f1p["iobj"] = iobj.Normal(name.NO_ART)
      f3p["iobj"] = f1p["iobj"]
      m = msg.New(gstring.Sprintm(pointTransWithTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointTransWithTemplate, f1p))
    } else {
      pp.QWrite("Look, that syntax just isn't supported yet.")
    }
    
  } else {
    var f2p = map[string]interface{} { "subj":    f3p["subj"],
                                       "subj_pp": f3p["subj_pp"],
                                       "verb":    f3p["verb"],
                                       "dobj":    "you", }
    f1p["dobj"] = dobj.Normal(0)
    f3p["dobj"] = f1p["dobj"]
    
    if iobj == nil {
      m = msg.New(gstring.Sprintm(pointTransTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointTransTemplate, f1p))
      m.Add(dobj, gstring.Sprintm(pointTransTemplate, f2p))
      
    } else if prep == "with" {
      if !pp.Body().IsHolding(iobj) {
        pp.QWrite("You are not holding %s.", iobj.Normal(name.DEF_ART))
        return
      }
      iname := iobj.Normal(name.NO_ART)
      f1p["iobj"], f2p["iobj"], f3p["iobj"] = iname, iname, iname
      m = msg.New(gstring.Sprintm(pointTransWithTemplate, f3p))
      m.Add(pp, gstring.Sprintm(pointTransWithTemplate, f1p))
      m.Add(dobj, gstring.Sprintm(pointTransWithTemplate, f2p))
      
    } else {
      pp.QWrite("Look, that syntax just isn't supported yet.")
      return
    }
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}
      
