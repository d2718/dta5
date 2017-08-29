// name.go
//
// dta5 Name interface and methods
//
// updated 2017-08-01
//
// The Name interface represents all the "grammatical" ways in which the
// game can refer to a thing.Thing, including shorter and longer descriptions
// for appropriate situations, as well as pronouns. It also has a method for
// matching abbreviated descriptions (as would appear in a player-typed
// command, like GRE SH for "a loosely-knit green shirt") to names.
//
// So far there are only two implementations of the Name interface:
// the ProperName for PlayerChars (and possibly NPCs, if there ever are any),
// and the NormalName for everything else.
//
package name

import( "strings" )

// Controls details about how the requested version of a name be displayed:
//
//  0                 normally: "an antique brass key"
//  POSS            possessive: "an antique brass key's"
//  NO_ART  without an article: "antique brass key"
//  DEF_ART         with "the": "the antique brass key"
//
// They can be or'd together:
//  t.Short(name.POSS|name.DEF_ART) => "the key's"
//
type NameParam byte
const(  //NONE    NameParam = 0
        POSS    NameParam = 1
        NO_ART  NameParam = 2
        DEF_ART NameParam = 4
)

// This function is used internally to see if an or'd-together NameParam
// has a specific value or'd in.
func checkNP(val, param NameParam) bool {
  return (val & param) > 0
}

// Here is how a Name should satisfy the following function calls:
//
//  n.Full(0)   => "an antique brass key etched with spidery tracework"
//  n.Normal(0) => "an antique brass key"
//  n.Short(0)  => "a key" (not the appropriate article change)
//  n.Normal(name.POSS|name.NO_ART) => "antique brass key's"
//  n.Short(name.DEF_ART)           => "the key"
//  n.SubjPronoun()   => "it"
//  n.ObjPronoun()    => "it"
//  n.PossPronoun()   => "its"
//  n.ReflexPronoun() => "itself"
//  n.Match(["br", "k"])   => true
//  n.Match(["ke"])        => true
//  n.Match(["an", "key"]) => false
//
// The Name.Full() method is never intended to be called with the POSS
// NameParam. The results are considered "undefined".
//
type Name interface {
  Full(NameParam)   string
  Normal(NameParam) string
  Short(NameParam)  string
  SubjPronoun()     string
  ObjPronoun()      string
  PossPronoun()     string
  ReflexPronoun()   string
  Match([]string)   bool
}

// The Article type is intended to be used internally to specify which
// article should be used under what circumstances.
//
type Article byte
const(  AA Article = iota // "a red student's notebook" / "a notebook"
        AN                // "an eerily-painted icon" / "an icon"
        AAN               // "a snowy white igloo" / "an igloo"
        ANA               // "an iron cutlass" / "a cutlass"
        SOME              // "some rancid burbling muck" / "some muck"
        THE               // "the solid black altar of Set" / "the altar of Set"
        NONE              // "Jane Bunyan the Lumberjill" / "Jane"
)

// The NormalName is intended to represent textual ways to refer to most
// objects in the game--anything that doesn't require a proper noun (and
// maybe even some of those that do). Some examples:
//  an unmatched sock / an unmatched sock / a sock
//  a door / a door / a door
//  a Valyrian steel dagger etched with scaly shapes /
//    a Valyrian steel dagger / a dagger
//  some patched and frayed trousers / some patched and frayed trousers /
//    some trousers
//  the eastern gate of the city / the eastern gate / the gate
//  a hideous stone gargoyle perched atop the eastern gate /
//    / a hideous stone gargoyle / a gargoyle
//
type NormalName struct {
  Article
  Noun       string
  Modifiers  []string
  PrepPhrase string
  genDropsS  bool
}

// Create a new NormalName struct and return a pointer to it.
//
// artAdjNoun is a single string that works as a kind of shorthand for the
// specification of the Article, the series of modifying words, and the
// noun (which always comes last). Some examples:
//  "a big black book"
//  "an eerily-pained icon"
//  "a/an brass-hilted estoc"
//  "an/a iron-hilted rapier"
//  "some fine-grained sand"
//  "the lone door"
//
// prep is for if a Thing should have a "prepositional phrase" after its
// noun in the Full()-length version of its name (for example):
// "a porcelain plate with images of cats painted on it"
// "some foul-smelling muck that burbles with a sickening sound"
//
// isPlural is just that -- this is used to determine whether the post-
// apostrophe 's' needs to be dropped in the possessive case.
//
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
  case "an", "anan":
    art = AN
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

// NormalName implements the name.Name interface.
//
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
    case AN, ANA:
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
    case AN, ANA:
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
    case AA, ANA:
      chunks = append(chunks, "a")
    case AN, AAN:
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

func (n NormalName) SubjPronoun()   string { return "it" }
func (n NormalName) ObjPronoun()    string { return "it" }
func (n NormalName) PossPronoun()   string { return "its" }
func (n NormalName) ReflexPronoun() string { return "itself" }

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

// Use by the saving process, this function turns the article, modifiers, and
// noun back into a single string fit to pass as the artAdjNoun parameter of
// the NewNormal() function.
//
func (n NormalName) ToSaveString() string {
  var a string
  switch n.Article {
  case AA:
    a = "a"
  case AN:
    a = "an"
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

// Because ProperNames refer to people, each needs a Gender associated with
// it. This is mainly for returning the appropriate pronouns.
//
type Gender byte
const(  IT    Gender = 0  // neuter singular
        THEY  Gender = 1  // plural
        HE    Gender = 2  // masculine singular
        SHE   Gender = 3  // feminine singular
)

// A ProperName is intended to name a person (specifically a PlayerChar). It
// responds thus:
//
//  n.Full(0)   => "High Imperator Bob The Magnificent"
//  n.Normal(0) => "High Imperator Bob"
//  n.Short(0)  => "Bob"
//
// They never include articles (this isn't Ancient Greek), so they ignore the
// passed NameParam value.
//
type ProperName struct {
  Title string
  First string
  Rest string
  Gender
}

// Unsurprisingly, ProperName implements the Name interface.
//
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

func (n ProperName) ReflexPronoun() string {
  switch n.Gender {
  case HE:
    return "himself"
  case SHE:
    return "herself"
  case THEY:
    return "themselves"
  default:
    return "itself"
  }
}

// The only part of a ProperName that can Match is the actual "first name"
// part.
//
func (n ProperName) Match(toks []string) bool {
  if len(toks) == 1 {
    if thisStartsThat(toks[0], n.First) {
      return true
    }
  }
  return false
}
  
  
