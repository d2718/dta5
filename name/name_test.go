// name_test.go
//
// testing dta5/name
//
// updated 2017-07-15
//
package name

import( "strings";
        "testing"
)

func eqNormalName(n, m *NormalName) bool {
  if n.Article != m.Article {
    return false
  } else if n.Noun != m.Noun {
    return false
  } else if n.PrepPhrase != m.PrepPhrase {
    return false
  } else if n.genDropsS != m.genDropsS {
    return false
  } else if len(n.Modifiers) != len(m.Modifiers) {
    return false
  } else {
    for n, x := range n.Modifiers {
      if x != m.Modifiers[n] {
        return false
      }
    }
  }
  return true
}

var nnames = []*NormalName {
  &NormalName{  Article: AA, Noun: "frog", Modifiers: []string{"blue", },
                PrepPhrase: "with feathery wings", genDropsS: false, },
  &NormalName{  Article: AA, Noun: "crab",
                Modifiers: []string{"cantankerous", "clawed", },
                PrepPhrase: "", genDropsS: false, },
  &NormalName{  Article: AAN, Noun: "frog", Modifiers: []string{"feathery", "orange", },
                PrepPhrase: "", genDropsS: false, },
  &NormalName{  Article: THE, Noun: "Truth", Modifiers: []string{"Universal", },
                PrepPhrase: "", genDropsS: false, },
  &NormalName{  Article: SOME, Noun: "misconceptions", Modifiers: []string{"common", },
                PrepPhrase: "", genDropsS: true, },
  &NormalName{  Article: SOME, Noun: "metal", Modifiers: []string{"rusted", },
                PrepPhrase: "", genDropsS: false, },
}

type new_normal_pair struct {
  name_str string
  prep_str string
  pl       bool
  output *NormalName
}

var new_normal_pairs = []new_normal_pair {
  { "a blue frog", "with feathery wings", false, nnames[0], },
  { "aa cantankerous clawed crab", "",    false, nnames[1], },
  { "a/an feathery orange frog",   "",    false, nnames[2], },
  { "The Universal Truth",         "",    false, nnames[3], },
  { "some common misconceptions",  "",    true,  nnames[4], },
  { "some rusted metal",           "",    false, nnames[5], },
}

type match_pair struct {
  input_str string
  target    Name
  result    bool
}

var normal_match_pairs = []match_pair {
  { "frog",               nnames[2], true, },
  { "f",                  nnames[2], true, },
  { "orange frog",        nnames[2], true, },
  { "or fr",              nnames[2], true, },
  { "oRaN fRoG",          nnames[2], true, },
  { "",                   nnames[2], false, },
  { "blue",               nnames[2], false, },
  { "F O F",              nnames[2], true, },
  { "rog",                nnames[2], false, },
  { "feath or fro",       nnames[2], true, },
  { "feath or frg",       nnames[2], false, },
  { "feathery blue frog", nnames[2], false, },
}

type short_norm_long_pair struct {
  the_name Name
  params   NameParam
  short    string
  normal   string
  full     string
}

var norm_snl_pairs = []short_norm_long_pair {
  { nnames[0], 0,         "a frog", "a blue frog",
                          "a blue frog with feathery wings", },
  { nnames[0], NO_ART,    "frog", "blue frog",
                          "blue frog with feathery wings", },
  { nnames[0], DEF_ART,   "the frog", "the blue frog",
                          "the blue frog with feathery wings", },
  { nnames[0], POSS,      "a frog's", "a blue frog's",
                          "a blue frog's with feathery wings", },
  { nnames[0], POSS|DEF_ART,  "the frog's", "the blue frog's", 
                              "the blue frog's with feathery wings", },
  { nnames[4], 0,         "some misconceptions", "some common misconceptions",
                          "some common misconceptions", },
  { nnames[4], POSS|DEF_ART,  "the misconceptions'", "the common misconceptions'",
                              "the common misconceptions'", },
}

func TestNormalName(t *testing.T) {
  for _, p := range new_normal_pairs {
    np := NewNormal(p.name_str, p.prep_str, p.pl)
    if !eqNormalName(np, p.output) {
      t.Errorf("%q, %q, %v => %v\nexpected %v\n",
                p.name_str, p.prep_str, p.pl, *np, *(p.output))
    }
  }
  
  for _, p := range normal_match_pairs {
    tokens := strings.Fields(p.input_str)
    res := p.target.Match(tokens)
    if res != p.result {
      t.Errorf("%q Match() %v => %v\nexpected %v\n",
                p.input_str, p.target, res, p.result)
    }
  }
  
  for _, p := range norm_snl_pairs {
    s := p.the_name.Short(p.params)
    n := p.the_name.Normal(p.params)
    f := p.the_name.Full(p.params)
    if s != p.short {
      t.Errorf("Short(%v) => %q\nexpected %q\n", p.params, s, p.short)
    }
    if n != p.normal {
      t.Errorf("Normal(%v) => %q\nexpected %q\n", p.params, n, p.normal)
    }
    if f != p.full {
      t.Errorf("Full(%v) => %q\nexpected %q\n", p.params, f, p.full)
    }
  }
}

var pnames = []*ProperName {
  &ProperName{ Title: "Lord", First: "Xanthias", Rest: "The Red", },
  &ProperName{ Title: "Dr.", First: "Walter", Rest: "E. Hill, Ph.D.", },
  &ProperName{ First: "Phil", },
}

var proper_match_pairs = []match_pair {
  { "xanth",        pnames[0], true, },
  { "Xan",          pnames[0], true, },
  { "lord xan",     pnames[0], false, },
  { "XANthIAs",     pnames[0], true, },
  { "",             pnames[0], false, },
  { "asDf",         pnames[0], false, },
  { "xanth",        pnames[1], false, },
  { "walt",         pnames[1], true, },
  { "walt, Ph.D.",  pnames[1], false, },
}

var proper_snl_pairs = []short_norm_long_pair {
  { pnames[0], 0,       "Xanthias", "Lord Xanthias", "Lord Xanthias The Red", },
  { pnames[0], POSS,    "Xanthias's", "Lord Xanthias's", "Lord Xanthias The Red", },
  { pnames[0], NO_ART,  "Xanthias", "Lord Xanthias", "Lord Xanthias The Red", },
  { pnames[2], 0,       "Phil", "Phil", "Phil", },
  { pnames[2], POSS,    "Phil's", "Phil's", "Phil", },
  { pnames[2], POSS|DEF_ART, "Phil's", "Phil's", "Phil", },
}

func TestProperName(t *testing.T) {
  for _, p := range proper_match_pairs {
    tokens := strings.Fields(p.input_str)
    res := p.target.Match(tokens)
    if res != p.result {
      t.Errorf("%q Match() %v => %v\nexpected %v\n",
                p.input_str, p.target, res, p.result)
    }
  }
  
  for _, p := range proper_snl_pairs {
    s := p.the_name.Short(p.params)
    n := p.the_name.Normal(p.params)
    f := p.the_name.Full(p.params)
    if s != p.short {
      t.Errorf("Short(%v) => %q\nexpected %q\n", p.params, s, p.short)
    }
    if n != p.normal {
      t.Errorf("Normal(%v) => %q\nexpected %q\n", p.params, n, p.normal)
    }
    if f != p.full {
      t.Errorf("Full(%v) => %q\nexpected %q\n", p.params, f, p.full)
    }
  }
}

