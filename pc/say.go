// say.go
//
// dta5 PlayerChar speech
//
// updated 2017-08-06
//
package pc

import( "strings";
        "github.com/delicb/gstring";
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

const sayTemplate   = `{subj} {verb}, "{text}{punct}"`
const sayToTemplate = `{subj} {verb} {targ}, "{text}{punct}"`

func DoSay(pp *PlayerChar, verb string, dobj thing.Thing,
           prep string, iobj thing.Thing, text string) {
  
  var toks []string
  
  if (text[0] == '"') || (text[0] == '\'') {
    toks = strings.Fields(text[1:])
  } else {
    toks = strings.Fields(text)[1:]
  }
  
  if len(toks) == 0 {
    pp.QWrite("Say what now?")
    return
  }
  
  var tgt_toks []string = make([]string, 0, 0)
  
  for n, tok := range toks {
    if tok[0] == '@' {
      if len(tok) > 1 {
        for _, x := range toks[n:] {
          tgt_toks = append(tgt_toks, strings.ToLower(x))
        }
        tgt_toks[0] = tgt_toks[0][1:]
        toks = toks[:n]
        break
      }
    }
  }
  
  var obj thing.Thing = nil
  if len(tgt_toks) > 0 {
    obj = pp.FindLikeSay(tgt_toks)
  }
  
  if len(toks) < 1 {
    if obj == nil {
      pp.QWrite("Say what now?")
      return
    } else {
      pp.QWrite("Say what to %s now?", obj.Normal(name.DEF_ART))
      return
    }
  }
  
  toks[0] = util.Cap(toks[0])
  wordz := []rune(strings.Join(toks, " "))
  punct_char := wordz[len(wordz)-1]
  
  var f1p = make(map[string]interface{})
  var f3p = make(map[string]interface{})
  var m *msg.Message
  
  f1p["subj"] = "You"
  f3p["subj"] = util.Cap(pp.Normal(0))
  f1p["text"] = string(wordz)
  f3p["text"] = f1p["text"]
  f1p["punct"], f3p["punct"] = "", ""
  
  if obj == nil {
    switch punct_char {
    case '.':
      f1p["verb"], f3p["verb"]  = "say", "says"

    case '!':
      f1p["verb"], f3p["verb"] = "exclaim", "exclaims"
    case '?':
      f1p["verb"], f3p["verb"] = "ask", "asks"
    default:
      f1p["verb"], f3p["verb"] = "say", "says"
      f1p["punct"], f3p["punct"] = ".", "."
    }
    m = msg.New("speech", gstring.Sprintm(sayTemplate, f3p))
    m.Add(pp, "speech", gstring.Sprintm(sayTemplate, f1p))
  } else if obj == pp {
    switch punct_char {
    case '.':
      f1p["verb"], f3p["verb"]  = "say to", "says to"

    case '!':
      f1p["verb"], f3p["verb"] = "exclaim to", "exclaims to"
    case '?':
      f1p["verb"], f3p["verb"] = "ask", "asks"
    default:
      f1p["verb"], f3p["verb"] = "say to", "says to"
      f1p["punct"], f3p["punct"] = ".", "."
    }
    f1p["targ"], f3p["targ"] = "yourself", pp.ReflexPronoun()
    m = msg.New("speech", gstring.Sprintm(sayToTemplate, f3p))
    m.Add(pp, "speech", gstring.Sprintm(sayToTemplate, f1p))
  } else {
    var f2p = map[string]interface{} {
      "subj":  f3p["subj"],
      "text":  f3p["text"],
      "punct": f3p["punct"],
      "targ":  "you",
    }
    f1p["targ"] = obj.Normal(0)
    f3p["targ"] = f1p["targ"]
    
    switch punct_char {
    case '.':
      f1p["verb"], f2p["verb"], f3p["verb"]  = "say to", "says to", "says to"
    case '!':
      f1p["verb"], f2p["verb"], f3p["verb"] = "exclaim", "exclaims to", "exclaims to"
      
    case '?':
      f1p["verb"], f2p["verb"], f3p["verb"] = "ask", "asks", "asks"
    default:
      f1p["verb"], f2p["verb"], f3p["verb"] = "say to", "says to", "says to"
      f1p["punct"], f2p["punct"], f3p["punct"] = ".", ".", "."
    }

    m = msg.New("speech", gstring.Sprintm(sayToTemplate, f3p))
    m.Add(pp, "speech", gstring.Sprintm(sayToTemplate, f1p))
    m.Add(obj, "speech", gstring.Sprintm(sayToTemplate, f2p))
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}
