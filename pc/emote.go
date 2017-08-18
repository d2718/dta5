// emote.go
//
// dta5 PlayerChar "directional emote" verbs
//
// updated 2017-08-09
//
package pc

import( "fmt";
        "github.com/delicb/gstring";
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

type verbConj struct {
  p1 string
  p3 string
}

var emoteVerbFilter = map[string]verbConj {
  "nod":    verbConj{ "nod your head", "nods his head" },
  "raise":  verbConj{ "raise your eyebrows", "raises {subj_pp} eyebrows" },
  "shake":  verbConj{ "shake your head", "shakes {subj_pp} head" },
  "snap":   verbConj{ "snap your fingers", "snaps his fingers" },
}

var dirDirMsgs = map[room.NavDir]string {
  room.N:   "northward",
  room.NE:  "northeastward",
  room.E:   "eastward",
  room.SE:  "southeastward",
  room.S:   "southward",
  room.SW:  "southwestward",
  room.W:   "westward",
  room.NW:  "northwestward",
  room.UP:  "upward",
  room.DOWN:"downward",
  room.OUT: "out",
  -1:       "forward",
  -2:       "backward",
  -3:       "to {subj_pp} left",
  -4:       "to {subj_pp} right",
}

var dirDefaultTemplate = "{subj} {verb} {dir}."
var dirGlanceTemplate = "{subj} {glance} {dir} and {verb}."

var dirTemplates = map[string]string {
  "chuckle":  dirGlanceTemplate,
  "frown":    dirGlanceTemplate,
  "raise":    dirGlanceTemplate,
  "shake":    dirGlanceTemplate,
  "shrug":    dirGlanceTemplate,
  "sigh":     dirGlanceTemplate,
  "snap":     dirGlanceTemplate,
  "sneer":    dirGlanceTemplate,
  "snicker":  dirGlanceTemplate,
}

func DoEmoteDir(pp *PlayerChar, verb string, dirNum room.NavDir) {
  var f1p = map[string]interface{} {  "subj_pp": "your",
                                      "subj": "you",
                                      "glance": "glance", }
  var f3p = map[string]interface{} {  "subj_pp": pp.PossPronoun(),
                                      "subj": pp.Normal(0),
                                      "glance": "glances", }
  var dir1p = gstring.Sprintm(dirDirMsgs[dirNum], f1p)
  var dir3p = gstring.Sprintm(dirDirMsgs[dirNum], f3p)
  f1p["dir"] = dir1p
  f3p["dir"] = dir3p
  var v verbConj
  var ok  bool
  if tv, ok := emoteVerbFilter[verb]; ok {
    v = verbConj{
      p1: gstring.Sprintm(tv.p1, f1p),
      p3: gstring.Sprintm(tv.p3, f3p),
    }
  } else {
    v = verbConj{ p1: verb, p3: verb + "s", }
  }
  f1p["verb"] = v.p1
  f3p["verb"] = v.p3
  
  var templ string
  if templ, ok = dirTemplates[verb]; !ok {
    templ = dirDefaultTemplate
  }
  
  m := msg.New("txt", util.Cap(gstring.Sprintm(templ, f3p)))
  m.Add(pp, "txt", util.Cap(gstring.Sprintm(templ, f1p)))
  
  pp.where.Place.(*room.Room).Deliver(m)

}

const emoteIntransTemplate = "{subj} {verb}."
const emoteReflexTemplate  = "{subj} {verb} {sp_prep} {subj_rp}."
const emoteIndOnlyTemplate = "{subj} {glance} {prep} {iobj} and {verb}."
const emoteDirOnlyTemplate = "{subj} {verb} {sp_prep} {dobj}."
const emoteBothTemplate    = "{subj} {verb} {sp_prep} {dobj} {prep} {iobj}."

var emoteDefaultSpPrep = "at"
var emoteSpPreps = map[string]string {
  "lean": "toward",
}

// type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

func DoEmote(pp *PlayerChar, verb string, dobj thing.Thing,
             prep string, iobj thing.Thing, text string) {
  
  var f1p = map[string]interface{} { "subj": "you",
                                     "subj_pp": "your", }
  var f3p = map[string]interface{} { "subj": pp.Normal(0),
                                     "subj_pp": pp.PossPronoun(), }
  var v verbConj
  if tv, ok := emoteVerbFilter[verb]; ok {
    v = verbConj{
      p1: gstring.Sprintm(tv.p1, f1p),
      p3: gstring.Sprintm(tv.p3, f3p),
    }
  } else {
    v = verbConj{ p1: verb, p3: verb + "s", }
  }
  f1p["verb"] = v.p1
  f3p["verb"] = v.p3
  
  var m *msg.Message
  
  if iobj == nil {
    if dobj == nil {
      m = msg.New("txt", util.Cap(gstring.Sprintm(emoteIntransTemplate, f3p)))
      m.Add(pp, "txt", util.Cap(gstring.Sprintm(emoteIntransTemplate, f1p)))
      
    } else if dobj == pp {
      f1p["subj_rp"] = "yourself"
      f3p["subj_rp"] = pp.ReflexPronoun()
      var sp_prep string
      var ok bool
      if sp_prep, ok = emoteSpPreps[verb]; !ok {
        sp_prep = emoteDefaultSpPrep
      }
      f1p["sp_prep"] = sp_prep
      f3p["sp_prep"] = sp_prep
      m = msg.New("txt", util.Cap(gstring.Sprintm(emoteReflexTemplate, f3p)))
      m.Add(pp, "txt", util.Cap(gstring.Sprintm(emoteReflexTemplate, f1p)))
      
    } else {
      f2p := map[string]interface{} {"subj": pp.Normal(0),
                                     "subj_pp": pp.PossPronoun(),
                                     "verb": f3p["verb"],
                                     "dobj": "you", }
      if pp.Inventory.Contains(dobj) {
        f1p["dobj"] = fmt.Sprintf("your %s", dobj.Normal(name.NO_ART))
        f3p["dobj"] = fmt.Sprintf("%s %s", pp.PossPronoun(), dobj.Normal(name.NO_ART))
      } else {
        f1p["dobj"] = dobj.Normal(0)
        f3p["dobj"] = f1p["dobj"]
      }
      var sp_prep string
      var ok bool
      if sp_prep, ok = emoteSpPreps[verb]; !ok {
        sp_prep = emoteDefaultSpPrep
      }
      f1p["sp_prep"] = sp_prep
      f2p["sp_prep"] = sp_prep
      f3p["sp_prep"] = sp_prep
      
      m = msg.New("txt", util.Cap(gstring.Sprintm(emoteDirOnlyTemplate, f3p)))
      m.Add(pp, "txt", util.Cap(gstring.Sprintm(emoteDirOnlyTemplate, f1p)))
      m.Add(dobj, "txt", util.Cap(gstring.Sprintm(emoteDirOnlyTemplate, f2p)))
    }
    
  } else {
    if prep == "on" {
      f1p["prep"] = "on top of"
      f3p["prep"] = "on top of"
    } else {
      f1p["prep"] = prep
      f3p["prep"] = prep
    }
    
    if pp.Inventory.Contains(iobj) {
      f1p["iobj"] = fmt.Sprintf("your %s", iobj.Normal(name.NO_ART))
      f3p["iobj"] = fmt.Sprintf("%s %s", pp.PossPronoun(), iobj.Normal(name.NO_ART))
    } else {
      f1p["iobj"] = iobj.Normal(0)
      f3p["iobj"] = f1p["iobj"]
    }
    
    if dobj == nil {
      f1p["glance"] = "glance"
      f3p["glance"] = "glances"
      
      m = msg.New("txt", util.Cap(gstring.Sprintm(emoteIndOnlyTemplate, f3p)))
      m.Add(pp, "txt", util.Cap(gstring.Sprintm(emoteIndOnlyTemplate, f1p)))
      
    } else {
      f1p["dobj"] = dobj.Normal(0)
      if prep == "on" {
        f3p["dobj"] = f1p["dobj"]
      } else {
        f3p["dobj"] = "something"
      }
      var sp_prep string
      var ok bool
      if sp_prep, ok = emoteSpPreps[verb]; !ok {
        sp_prep = emoteDefaultSpPrep
      }
      f1p["sp_prep"] = sp_prep
      f3p["sp_prep"] = sp_prep
      
      m = msg.New("txt", util.Cap(gstring.Sprintm(emoteBothTemplate, f3p)))
      m.Add(pp, "txt", util.Cap(gstring.Sprintm(emoteBothTemplate, f1p)))
    }
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}
