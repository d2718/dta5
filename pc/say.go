// say.go
//
// dta5 PlayerChar speech
//
// updated 2017-08-06
//
package pc

import( "strings"
        "dta5/msg"; "dta5/name"; "dta5/room"; "dta5/thing"; "dta5/util";
)

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
  var m *msg.Message
  
  switch punct_char {
  case '.':
    if obj == nil {
      m = msg.New("%s says, \"%s\"", util.Cap(pp.Normal(0)), string(wordz))
      m.Add(pp, "You say, \"%s\"", string(wordz))
    } else {
      m = msg.New("%s says to %s, \"%s\"", util.Cap(pp.Normal(0)),
                  obj.Normal(0), string(wordz))
      m.Add(obj, "%s says to you, \"%s\"", util.Cap(pp.Normal(0)),
                  string(wordz))
      m.Add(pp, "You say to %s, \"%s\"", obj.Normal(0), string(wordz))
    }
  case '!':
    if obj == nil {
      m = msg.New("%s exclaims, \"%s\"", util.Cap(pp.Normal(0)), string(wordz))
      m.Add(pp, "You exclaim, \"%s\"", string(wordz))
    } else {
      m = msg.New("%s exclaims to %s, \"%s\"", util.Cap(pp.Normal(0)),
                  obj.Normal(0), string(wordz))
      m.Add(obj, "%s exclaims to you, \"%s\"", util.Cap(pp.Normal(0)),
                  string(wordz))
      m.Add(pp, "You exclaim to %s, \"%s\"", obj.Normal(0), string(wordz))
    }
  case '?':
    if obj == nil {
      m = msg.New("%s asks, \"%s\"", util.Cap(pp.Normal(0)), string(wordz))
      m.Add(pp, "You ask, \"%s\"", string(wordz))
    } else {
      m = msg.New("%s asks %s, \"%s\"", util.Cap(pp.Normal(0)),
                  obj.Normal(0), string(wordz))
      m.Add(obj, "%s asks you, \"%s\"", util.Cap(pp.Normal(0)),
                  string(wordz))
      m.Add(pp, "You ask %s, \"%s\"", obj.Normal(0), string(wordz))
    }
  default:
    if obj == nil {
      m = msg.New("%s says, \"%s.\"", util.Cap(pp.Normal(0)), string(wordz))
      m.Add(pp, "You say, \"%s\"", string(wordz))
    } else {
      m = msg.New("%s says to %s, \"%s.\"", util.Cap(pp.Normal(0)),
                  obj.Normal(0), string(wordz))
      m.Add(obj, "%s says to you, \"%s.\"", util.Cap(pp.Normal(0)),
                  string(wordz))
      m.Add(pp, "You say to %s, \"%s.\"", obj.Normal(0), string(wordz))
    }
  }
  
  pp.where.Place.(*room.Room).Deliver(m)
}
