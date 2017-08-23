// parse.go
//
// the dta5 player character command parser
//
// updated 2017-08-14
//
package pc

import( "strings";
        "dta5/msg";
        "dta5/name"; "dta5/room"; "dta5/scripts"; "dta5/thing"; "dta5/util";
)

var parseVerbs []string = []string{
  "close", "drop", "examine", "exits", "get", "go", "help", "inventory", 
  "look", "lock", "open", "put", "remove", "say", "swap", "take",
  "unlock", "wear",
  
  // point/wave (pc/point.go)
  
  "point", "wave",
  
  // directional emote verbs (pc/emote.go)
  "blink", "chuckle", "gaze", "frown", "glance", "grin", "lean",
  "nod", "raise", "shake", "shrug", "sigh", "snap", "sneer",
  "snicker", "squint", "stare", "wink",
}

var verbTranslation map[string]string = map[string]string {
  "take": "get",
  "drop": "put",
}

var parsePreps map[string]byte = map[string]byte {
  "behind": thing.BEHIND,
  "in": thing.IN,
  "on": thing.ON,
  "under": thing.UNDER,
  "with": 127,
  //"at": 126,
}

var ordinals map[string]int = map[string]int {
  "first": 0, "second": 1, "third": 2, "fourth": 3, "fifth": 4, "sixth": 5,
  "seventh": 6, "eighth": 7, "ninth": 8, "tenth": 9, "eleventh": 10,
  "twelfth": 11,
  "other": 1, // just an alias for "second"
}

type ParseFunc func(*PlayerChar, string, []string, string)
type DoFunc func(*PlayerChar, string, thing.Thing, string, thing.Thing, string)

var parseDispatch map[string]ParseFunc = map[string]ParseFunc {
  "close":      ParseLikeLook,
  "examine":    ParseLikePut,
  "exits":      ParseIntransitive,
  "get":        ParseLikeLook,
  "go":         ParseGo,
  "help":       ParseIntransitive,
  "inventory":  ParseIntransitive,
  "lock":       ParseLikeLock,
  "look":       ParseLikeLook,
  "open":       ParseLikeLook,
  "put":        ParseLikePut,
  "remove":     ParseLikePut,
  "say":        ParseIntransitive,
  "swap":       ParseIntransitive,
  "unlock":     ParseLikeLock,
  "wear":       ParseLikePut,
  
  // point/wave
  
  "point":      ParseLikePoint,
  "wave":       ParseLikePoint,
  
  // directional emote verbs
  
  "blink":      ParseEmote,
  "chuckle":    ParseEmote,
  "frown":      ParseEmote,
  "gaze":       ParseEmote,
  "glance":     ParseEmote,
  "grin":       ParseEmote,
  "lean":       ParseEmote,
  "nod":        ParseEmote,
  "raise":      ParseEmote,
  "shake":      ParseEmote,
  "shrug":      ParseEmote,
  "sigh":       ParseEmote,
  "snap":       ParseEmote,
  "sneer":      ParseEmote,
  "snicker":    ParseEmote,
  "squint":     ParseEmote,
  "stare":      ParseEmote,
  "wink":       ParseEmote,
}

var doDispatch map[string]DoFunc = map[string]DoFunc {
  "close":      DoClose,
  "examine":    DoExamine,
  "exits":      DoExits,
  "get":        DoGet,
  "go":         DoMove,
  "help":       DoHelp,
  "inventory":  DoInventory,
  "lock":       DoLock,
  "look":       DoLook,
  "open":       DoOpen,
  "put":        DoPut,
  "remove":     DoRemove,
  "say":        DoSay,
  "swap":       DoSwap,
  "unlock":     DoLock, // this is correct
  "wear":       DoWear,
  
  // point/wave
  
  "point":      DoPoint,
  "wave":       DoPoint,
  
  // directional emote verbs

  "blink":      DoEmote,
  "chuckle":    DoEmote,
  "frown":      DoEmote,
  "gaze":       DoEmote,
  "glance":     DoEmote,
  "grin":       DoEmote,
  "lean":       DoEmote,
  "nod":        DoEmote,
  "raise":      DoEmote,
  "shake":      DoEmote,
  "shrug":      DoEmote,
  "sigh":       DoEmote,
  "snap":       DoEmote,
  "sneer":      DoEmote,
  "snicker":    DoEmote,
  "squint":     DoEmote,
  "stare":      DoEmote,
  "wink":       DoEmote,
}

var cardDirs map[string]room.NavDir = map[string]room.NavDir {
  "n":  room.N,     "north": room.N,
  "ne": room.NE,    "northeast": room.NE,
  "e":  room.E,     "east": room.E,
  "se": room.SE,    "southeast": room.SE,
  "s":  room.S,     "south": room.S,
  "sw": room.SW,    "southwest": room.SW,
  "w":  room.W,     "west": room.W,
  "nw": room.NW,    "northwest": room.NW,
  "u":  room.UP,    "up": room.UP,
  "d":  room.DOWN,  "down": room.DOWN,
  "o":  room.OUT,   "out": room.OUT, }

var cardDirNames map[room.NavDir]string = map[room.NavDir]string {
  room.N: "north", room.NE: "northeast", room.E: "east",
  room.SE: "southeast", room.S: "south", room.SW: "southwest",
  room.W: "west", room.NW: "northwest", room.UP: "up", room.DOWN: "down",
  room.OUT: "out", }

var emoteDirs map[string]room.NavDir = map[string]room.NavDir {
  "forward":  room.NavDir(-1),
  "back":     room.NavDir(-2),
  "left":     room.NavDir(-3),
  "right":    room.NavDir(-4),
}

func thisStartsThat(this, that string) bool {
  try, targ := []rune(this), []rune(that)
  if len(try) > len(targ) {
    return false
  }
  for n, r := range try {
    if r != targ[n] {
      return false
    }
  }
  return true
}

func matchVerb(token string) string {
  for _, v := range parseVerbs {
    if thisStartsThat(token, v) {
      return v
    }
  }
  return ""
}

func (pp *PlayerChar) Parse(cmd string) error {
  
  pp.Send(msg.Env{Type: "echo", Text: cmd})
  
  toks := strings.Fields(strings.ToLower(cmd))
  if len(toks) == 0 {
    pp.QWrite("Sorry, what?")
    return nil
  }
  
  // process shortcuts
  if (cmd[0] == '"') || (cmd[0] == '\'') {
    DoSay(pp, "", nil, "", nil, cmd)
    return nil
  }
  
  if len(toks) == 1 {
    if dir, ok := cardDirs[toks[0]]; ok {
      DoMoveDir(pp, dir)
      return nil
    }
    
    if toks[0] == "quit" {
      return pp.Logout("You have quit the game.")
    }
    
  }
  
  var t_verb string = toks[0]
  var verb string = ""
  var ok bool = false
  toks = toks[1:]
  
  for _, v := range parseVerbs {
    if thisStartsThat(t_verb, v) {
      verb = v
      break
    }
  }
  
  if t_verb, ok = verbTranslation[verb]; ok {
    verb = t_verb
  }
  
  if parse_func, ok := parseDispatch[verb]; ok {
    parse_func(pp, verb, toks, cmd)
  } else {
    pp.QWrite("You don't appear to know how to %q.", t_verb)
  }
  
  return nil
}

func FindInInventory(pp *PlayerChar, toks []string, ord int) (thing.Thing, int) {
  return pp.Inventory.Find(toks, ord)
}
func FindInSurroundingsFirst(pp *PlayerChar, toks []string, ord int) (thing.Thing, int) {
  loc := pp.where.Place.(*room.Room)
  
  t, remain := loc.Contents.Find(toks, ord)
  if t != nil {
    return t, 0
  }
  t, remain = loc.Scenery.Find(toks, remain)
  if t != nil {
    return t, 0
  }
  return pp.Inventory.Find(toks, remain)
}

func FindInInventoryFirst(pp *PlayerChar, toks []string, ord int) (thing.Thing, int) {
  t, remain := pp.Inventory.Find(toks, ord)
  if t != nil {
    return t, 0
  }
  
  loc := pp.where.Place.(*room.Room)
  t, remain = loc.Contents.Find(toks, remain)
  if t != nil {
    return t, 0
  }
  return loc.Scenery.Find(toks, remain)
}

func FindPlayerChar(pp *PlayerChar, toks []string, ord int) (thing.Thing, int) {
  loc_stuff := pp.where.Place.(*room.Room).Contents.Things
  
  var remain int = ord
  for _, t := range loc_stuff {
    switch tt := t.(type) {
    case *PlayerChar:
      if tt.Match(toks) {
        if remain == 0 {
          return tt, 0
        } else {
          remain--
        }
      }
    default:
      // don't do anything
    }
  }
  return nil, remain
}

func (pp *PlayerChar) FindLikeLook(toks []string) thing.Thing {
  find_func := FindInSurroundingsFirst
  
  if toks[0] == "my" {
    toks = toks[1:]
    find_func = FindInInventory
  }
  if len(toks) < 1 {
    return nil
    
  }
  
  var ord int = 0
  var has_ord bool = false
  ord, has_ord = ordinals[toks[0]]
  if has_ord {
    toks = toks[1:]
  }
  if len(toks) < 1 {
    return nil
  }
  
  t, _ := find_func(pp, toks, ord)
  return t
}

func (pp *PlayerChar) FindLikePut(toks []string) thing.Thing {
  find_func := FindInInventoryFirst
  
  if toks[0] == "my" {
    toks = toks[1:]
    find_func = FindInInventory
  }
  if len(toks) < 1 {
    return nil
  }
  
  var ord int = 0
  var has_ord bool = false
  ord, has_ord = ordinals[toks[0]]
  if has_ord {
    toks = toks[1:]
  }
  if len(toks) < 1 {
    return nil
  }
  
  t, _ := find_func(pp, toks, ord)
  return t
}

func (pp *PlayerChar) FindLikeSay(toks []string) thing.Thing {
  var ord int = 0
  var has_ord bool = false
  pc_find_toks := toks
  ord, has_ord = ordinals[toks[0]]
  if has_ord {
    pc_find_toks = toks[1:]
  }
  if len(pc_find_toks) < 1 {
    return nil
  }
  
  t, rem := FindPlayerChar(pp, pc_find_toks, ord)
  if t != nil {
    return t
  }
  
  if has_ord {
    t, rem = FindInSurroundingsFirst(pp, pc_find_toks, rem)
  } else {
    find_func := FindInSurroundingsFirst
    if toks[0] == "my" {
      toks = toks[1:]
      if len(toks) < 1 {
        return nil
      }
      find_func = FindInInventory
    }
    ord, has_ord = ordinals[toks[0]]
    if has_ord {
      toks = toks[1:]
      if len(toks) < 1 {
        return nil
      }
    }
    t, rem = find_func(pp, toks, ord)
  }
  return t
}

func FindInThingList(tl *thing.ThingList, toks []string) thing.Thing {
  var ord int = 0
  var has_ord bool = false
  ord, has_ord = ordinals[toks[0]]
  if has_ord {
    toks = toks[1:]
  }
  t, _ := tl.Find(toks, ord)
  return t
}

// ParseIntransitive() ignores everything but the verb.
//
func ParseIntransitive(subj *PlayerChar, verb string, toks []string, text string) {
  doDispatch[verb](subj, verb, nil, "", nil, text)
}

// ParseLikeLook() checks for the following options
//
//  * no objects
//  * verb dir.obj.
//  * verb prep ind.obj.
//  * verb dir.obj. [must be] prep ind.obj.
//    (If there is both a direct object and an indirect object, the direct
//     object must be in/on/etc. [according to the supplied preposition]
//     the indirect object.)
//
func ParseLikeLook(subj *PlayerChar, verb string, toks []string, text string) {
  var prep string
  var prep_idx int = -1
  var dobj_toks []string
  var iobj_toks []string
  var dobj, iobj thing.Thing = nil, nil
  
  var ok bool
  for n, w := range toks {
    _, ok = parsePreps[w]
    if ok {
       prep = w
       prep_idx = n
       break
    }
  }
  
  if prep == "with" {
    subj.QWrite("You cannot %v \"with\" something.", verb)
    return
  }
  
  if prep_idx > -1 {
    dobj_toks = toks[:prep_idx]
    iobj_toks = toks[prep_idx+1:]
  } else {
    dobj_toks = toks
  }
  
  if prep == "at" {
    if len(dobj_toks) > 0 {
      subj.QWrite("That doesn't make any sense.")
      return
    } else {
      dobj_toks = iobj_toks
      iobj_toks = nil
    }
  }
  
  if len(iobj_toks) > 0 {
    iobj = subj.FindLikeLook(iobj_toks)
    if iobj == nil {
      subj.QWrite("You can not see any %q here.",
                  strings.Join(iobj_toks, " "))
      return
    }
    
    switch t_iobj := iobj.(type) {
    case thing.Container:
      s := t_iobj.Side(parsePreps[prep])
      if s == nil {
        subj.QWrite("There is nothing %s %s.", prep, t_iobj.Normal(name.DEF_ART))
        return
      }
      if (prep == "in") && (t_iobj.IsOpen() == false) {
        subj.QWrite("%s is closed.", util.Cap(t_iobj.Normal(name.DEF_ART)))
        return
      }
      
      if len(dobj_toks) > 0 {
        dobj = FindInThingList(s, dobj_toks)
        if dobj == nil {
          subj.QWrite("You cannot find any \"%s\" %s %s.",
                      strings.Join(dobj_toks, " "), prep,
                      t_iobj.Normal(name.DEF_ART))
          return
        }
      }
    default:
      subj.QWrite("You cannot see %s %s.", prep, iobj.Normal(0))
      return
    }
  } else if len(dobj_toks) > 0 {
    dobj = subj.FindLikeLook(dobj_toks)
    if dobj == nil {
      subj.QWrite("You can not see any %q here.",
                  strings.Join(dobj_toks, " "))
      return
    }
  }
  
  if scripts.Check(subj, dobj, iobj, verb, prep, text) {
    doDispatch[verb](subj, verb, dobj, prep, iobj, text)
  }
  
}


// ParseLikePut() checks for the following options
//
//  * no objects
//  * dir.obj. only
//  * ind.obj. only
//  * both objects, but their relationship is unimportant
//
// ParseLikePut() gives priority to the player's own inventory when looking
// for direct objects. (This is maybe a bad idea, and might change.)
//
func ParseLikePut(subj *PlayerChar, verb string, toks []string, text string) {
  var prep string
  var prep_idx int = -1
  var dobj_toks []string
  var iobj_toks []string
  var dobj, iobj thing.Thing = nil, nil
  
  var ok bool
  for n, w := range toks {
    _, ok = parsePreps[w]
    if ok {
       prep = w
       prep_idx = n
       break
    }
  }

  if prep == "with" {
    subj.QWrite("You cannot %v \"with\" something.", verb)
    return
  }
  
  if prep_idx > -1 {
    dobj_toks = toks[:prep_idx]
    iobj_toks = toks[prep_idx+1:]
  } else {
    dobj_toks = toks
  }
  
  if prep == "at" {
    subj.QWrite("That doesn't make any sense.")
    return
  }
  
  if len(dobj_toks) > 0 {
    dobj = subj.FindLikePut(dobj_toks)
    if dobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(dobj_toks, " "))
      return
    }
  }
  
  if len(iobj_toks) > 0 {
    iobj = subj.FindLikeLook(iobj_toks)
    if iobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(iobj_toks, " "))
      return
    }
  }
  
  if scripts.Check(subj, dobj, iobj, verb, prep, text) {
    doDispatch[verb](subj, verb, dobj, prep, iobj, text)
  }
}

// ParseLikeLock() checks for the following options
//
//  * no objects
//  * dir.obj. only
//  * ind.obj. only
//  * both objects, but their relationship is unimportant
//
// ParseLikeLock() gives priority to the player's own inventory when looking
// for indirect objects. (This is maybe a bad idea, and might change.)
//
func ParseLikeLock(subj *PlayerChar, verb string, toks []string, text string) {
  var prep string
  var prep_idx int = -1
  var dobj_toks []string
  var iobj_toks []string
  var dobj, iobj thing.Thing = nil, nil
  
  var ok bool
  for n, w := range toks {
    _, ok = parsePreps[w]
    if ok {
       prep = w
       prep_idx = n
       break
    }
  }
  
  if prep_idx > -1 {
    dobj_toks = toks[:prep_idx]
    iobj_toks = toks[prep_idx+1:]
  } else {
    dobj_toks = toks
  }
  
  if len(dobj_toks) > 0 {
    dobj = subj.FindLikeLook(dobj_toks)
    if dobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(dobj_toks, " "))
      return
    }
  }
  
  if len(iobj_toks) > 0 {
    iobj = subj.FindLikePut(iobj_toks)
    if iobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(iobj_toks, " "))
      return
    }
  }
  
  if scripts.Check(subj, dobj, iobj, verb, prep, text) {
    doDispatch[verb](subj, verb, dobj, prep, iobj, text)
  }
}

func ParseEmote(subj *PlayerChar, verb string, toks []string, text string) {
  if len(toks) == 1 {
    dir := toks[0]
    if dir_num, ok := cardDirs[dir]; ok {
      DoEmoteDir(subj, verb, dir_num)
      return
    }
    if dir_num, ok := emoteDirs[dir]; ok {
      DoEmoteDir(subj, verb, dir_num)
      return
    }
  }
  
  ParseLikeLook(subj, verb, toks, text)
}

func ParseGo(subj *PlayerChar, verb string, toks []string, text string) {
  if len(toks) == 1 {
    dir := toks[0]
    if dir_num, ok := cardDirs[dir]; ok {
      DoMoveDir(subj, dir_num)
      return
    }
  }
  
  ParseLikeLook(subj, verb, toks, text)
}

func ParseLikePoint(subj *PlayerChar, verb string, toks []string, text string) {
  var prep string
  var prep_idx int = -1
  var dobj_toks []string
  var iobj_toks []string
  var dobj, iobj thing.Thing = nil, nil
  
  var ok bool
  for n, w := range toks {
    _, ok = parsePreps[w]
    if ok {
       prep = w
       prep_idx = n
       break
    }
  }
  
  if prep_idx > -1 {
    dobj_toks = toks[:prep_idx]
    iobj_toks = toks[prep_idx+1:]
  } else {
    dobj_toks = toks
  }
  
  if len(iobj_toks) > 0 {
    iobj = subj.FindLikePut(iobj_toks)
    if iobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(iobj_toks, " "))
      return
    }
  }
  
  if len(dobj_toks) == 1 {
    if dir_num, ok := cardDirs[dobj_toks[0]]; ok {
      DoPointDir(subj, verb, dir_num, iobj)
      return
    }
    
    if dir_num, ok := emoteDirs[dobj_toks[0]]; ok {
      DoPointDir(subj, verb, dir_num, iobj)
      return
    }
  }
  
  if len(dobj_toks) > 0 {
    dobj = subj.FindLikeLook(dobj_toks)
    if dobj == nil {
      subj.QWrite("You cannot see any %q here.",
                  strings.Join(dobj_toks, " "))
      return
    }
  }

  if scripts.Check(subj, dobj, iobj, verb, prep, text) {
    doDispatch[verb](subj, verb, dobj, prep, iobj, text)
  }
}
