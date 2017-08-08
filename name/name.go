// name.go
//
// dta5 Name struct and methods
//
// updated 2017-08-01
//
package name

import( "strings";
)

type NameParam byte
const(  //NONE    NameParam = 0
        POSS    NameParam = 1
        NO_ART  NameParam = 2
        DEF_ART NameParam = 4
)

func checkNP(val, param NameParam) bool {
  return (val & param) > 0
}

type Name interface {
  Full(NameParam)   string
  Normal(NameParam) string
  Short(NameParam)  string
  SubjPronoun()     string
  ObjPronoun()      string
  PossPronoun()     string
  Match([]string)   bool
}

type Article byte
const(  AA    Article = iota
        AAN
        ANA
        SOME
        THE
        NONE
)

type NormalName struct {
  Article
  Noun       string
  Modifiers  []string
  PrepPhrase string
  genDropsS  bool
}

func NewNormal(artAdjNoun, prep string, isPlural bool) *NormalName {
  chunks := strings.Fields(artAdjNoun)
  var gds bool = false
  if len(chunks) < 1 {
    return nil
  } else if len(chunks) == 1 {
    noun := chunks[0]
    if isPlural {
      if (strings.ToLower(noun))[len(noun)-1] == 's' {
        gds = true
      }
    }
    return &NormalName{ Article: NONE,
                        Noun: noun,
                        Modifiers: make([]string, 0, 0),
                        PrepPhrase: prep,
                        genDropsS: gds,
                      }
  }
  
  var adjs []string
  var newlen int = len(chunks)-1
  var noun = chunks[newlen]
  var art Article = NONE
  switch strings.ToLower(chunks[0]) {
  case "a", "aa", "a/a":
    art = AA
    adjs = chunks[1:newlen]
  case "aan", "a/an":
    art = AAN
    adjs = chunks[1:newlen]
  case "ana", "an/a":
    art = ANA
    adjs = chunks[1:newlen]
  case "some":
    art = SOME
    adjs = chunks[1:newlen]
  case "the":
    art = THE
    adjs = chunks[1:newlen]
  default:
    art = NONE
    adjs = chunks[:newlen]
  }
  
  if isPlural {
    if (strings.ToLower(noun))[len(noun)-1] == 's' {
      gds = true
    }
  }
  
  return &NormalName{ Article: art,
                      Noun: noun,
                      Modifiers: adjs,
                      PrepPhrase: prep,
                      genDropsS: gds,
                    }
}

// Not intended to be called with POSS>
func (n NormalName) Full(p NameParam) string {
  var chunks []string = make([]string, 0, 4)
  
  if checkNP(p, DEF_ART) {
    if n.Article != NONE {
      chunks = append(chunks, "the")
    }
  } else if !checkNP(p, NO_ART) {
    switch n.Article {
    case AA, AAN:
      chunks = append(chunks, "a")
    case ANA:
      chunks = append(chunks, "an")
    case SOME:
      chunks = append(chunks, "some")
    case THE:
      chunks = append(chunks, "the")
    }
  }
  
  if len(n.Modifiers) > 0 {
    chunks = append(chunks, strings.Join(n.Modifiers, " "))
  }
  
  if checkNP(p, POSS) {
    if n.genDropsS {
      chunks = append(chunks, n.Noun + "'")
    } else {
      chunks = append(chunks, n.Noun + "'s")
    }
  } else {
    chunks = append(chunks, n.Noun)
  }
  
  if (n.PrepPhrase != "") {
    chunks = append(chunks, n.PrepPhrase)
  }
  
  return strings.Join(chunks, " ")
}

func (n NormalName) Normal(p NameParam) string {
  var chunks []string = make([]string, 0, 3)
  
  if checkNP(p, DEF_ART) {
    if n.Article != NONE {
      chunks = append(chunks, "the")
    }
  } else if !checkNP(p, NO_ART) {
    switch n.Article {
    case AA, AAN:
      chunks = append(chunks, "a")
    case ANA:
      chunks = append(chunks, "an")
    case SOME:
      chunks = append(chunks, "some")
    case THE:
      chunks = append(chunks, "the")
    }
  }
  
  if len(n.Modifiers) > 0 {
    chunks = append(chunks, strings.Join(n.Modifiers, " "))
  }
  
  if checkNP(p, POSS) {
    if n.genDropsS {
      chunks = append(chunks, n.Noun + "'")
    } else {
      chunks = append(chunks, n.Noun + "'s")
    }
  } else {
    chunks = append(chunks, n.Noun)
  }
  
  return strings.Join(chunks, " ")
}

func (n NormalName) Short(p NameParam) string {
  var chunks []string = make([]string, 0, 2)

  if checkNP(p, DEF_ART) {
    if n.Article != NONE {
      chunks = append(chunks, "the")
    }
  } else if !checkNP(p, NO_ART) {
    switch n.Article {
    case AA, AAN:
      chunks = append(chunks, "a")
    case ANA:
      chunks = append(chunks, "an")
    case SOME:
      chunks = append(chunks, "some")
    case THE:
      chunks = append(chunks, "the")
    }
  }
  
  if checkNP(p, POSS) {
    if n.genDropsS {
      chunks = append(chunks, n.Noun + "'")
    } else {
      chunks = append(chunks, n.Noun + "'s")
    }
  } else {
    chunks = append(chunks, n.Noun)
  }
  
  return strings.Join(chunks, " ")
}

func (n NormalName) SubjPronoun() string { return "it" }
func (n NormalName) ObjPronoun()  string { return "it" }
func (n NormalName) PossPronoun() string { return "its" }

func thisStartsThat(this, that string) bool {
  var ilen, alen int = len(this), len(that)
  if ilen > alen {
    return false
  }
  
  var i []rune = []rune(strings.ToLower(this))
  var a []rune = []rune(strings.ToLower(that))
  for n := 0; n < ilen; n++ {
    if i[n] != a[n] {
      return false
    }
  }
  
  return true
}

func (n NormalName) Match(toks []string) bool {
  n_toks := len(toks)
  
  if n_toks < 1 {
    return false
  } else if n_toks > len(n.Modifiers) + 1 {
    return false
  }
  
  noun := toks[n_toks-1]
  adjs := toks[:n_toks-1]
  n_toks--
  
  if !thisStartsThat(noun, n.Noun) {
    return false
  }
  
  for i := 1; i <= n_toks; i++ {
    if !thisStartsThat(adjs[n_toks-i], n.Modifiers[len(n.Modifiers)-i]) {
      return false
    }
  }
  
  return true
}

func (n NormalName) ToSaveString() string {
  var a string
  switch n.Article {
  case AA:
    a = "a"
  case AAN:
    a = "a/an"
  case ANA:
    a = "an/a"
  case SOME:
    a = "some"
  case THE:
    a = "the"
  }
  
  cont := make([]string, 0, len(n.Modifiers)+2)
  cont = append(cont, a)
  for _, m := range n.Modifiers {
    cont = append(cont, m)
  }
  cont = append(cont, n.Noun)
  
  return strings.Join(cont, " ")
}

type Gender byte
const(  IT    Gender = 0
        THEY  Gender = 1
        HE    Gender = 2
        SHE   Gender = 3
)

type ProperName struct {
  Title string
  First string
  Rest string
  Gender
}

// Not intended to be called with POSS.
func (n ProperName) Full(p NameParam) string {
  var chunks []string = make([]string, 0, 3)
  
  if n.Title != "" {
    chunks = append(chunks, n.Title)
  }
  
  chunks = append(chunks, n.First)
  
  if n.Rest != "" {
    chunks = append(chunks, n.Rest)
  }
  
  return strings.Join(chunks, " ")
}

func (n ProperName) Normal(p NameParam) string {
  var chunks []string = make([]string, 0, 2)
  
  if n.Title != "" {
    chunks = append(chunks, n.Title)
  }
  
  if checkNP(p, POSS) {
    if n.Gender == THEY {
      chunks = append(chunks, n.First + "'")
    } else {
      chunks = append(chunks, n.First + "'s")
    }
  } else {
    chunks = append(chunks, n.First)
  }
  
  return strings.Join(chunks, " ")
}

func (n ProperName) Short(p NameParam) string {
  if checkNP(p, POSS) {
    if n.Gender == THEY {
      return n.First + "'"
    } else {
      return n.First + "'s"
    }
  } else {
    return n.First
  }
}

func (n ProperName) SubjPronoun() string {
  switch n.Gender {
  case HE:
    return "he"
  case SHE:
    return "she"
  case THEY:
    return "they"
  default:
    return "it"
  }
}

func (n ProperName) ObjPronoun() string {
  switch n.Gender {
  case HE:
    return "him"
  case SHE:
    return "her"
  case THEY:
    return "them"
  default:
    return "it"
  }
}

func (n ProperName) PossPronoun() string {
  switch n.Gender {
  case HE:
    return "his"
  case SHE:
    return "her"
  case THEY:
    return "their"
  default:
    return "its"
  }
}

func (n ProperName) Match(toks []string) bool {
  if len(toks) == 1 {
    if thisStartsThat(toks[0], n.First) {
      return true
    }
  }
  return false
}
  
  
