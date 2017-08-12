// load.go
//
// dta5 loading and world-building module
//
// 2017-08-11

// In data file, formats are
//
// [ type_string, ref, ... ]
//
// room.Room:
// ["room", "ref", "title", "nav targets"... ]
//
// thing.Item:
// ["item", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk ]
//
// thing.ItemContainer:
// ["itemc", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk, toggleable, open,
//           { "i": [in_mass, in_bulk], "o": [on_mass, on_bulk] ... } ]
//
// thing.Clothing:
// ["cloth", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk, "slot" ]
//
// thing.WornContainer:
// ["clothc", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk,
//            "slot", will_toggle, is_open, mass_held, bulk_held ]
//
// door.Doorway:
// ["dwy", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk,
//  WillToggle ]
//
// to populate a Room or Container
// ["pop", "ref", "side_string", "ref_list"... ]
//
// to add a MoodMessaging object
// ["mood", min_secs, max_secs, [ room_refs... ], [ messages...] ]
//
// to bind a script to a thing
// ["bind", "obj_ref", "verb", "script_tag" ]
//
// to use a world-building function from load/build
// ["build", "func_tag", args ... ]
//
// to load another file
// ["load", "relative filename" ]

//
package load

import( "encoding/json"; "fmt"; "os"; "path/filepath";
        "dta5/door"; "dta5/log"; "dta5/mood"; "dta5/ref";
        "dta5/room"; "dta5/scripts"; "dta5/thing";
        "dta5/load/build";
)

var WorldDir string

var sideMap map[rune]byte = map[rune]byte {
  's':  room.SCENERY,
  'c':  room.CONTENTS,
  'o':  thing.ON,
  'i':  thing.IN,
  'b':  thing.BEHIND,
  'u':  thing.UNDER,
}

func str2side(s string) byte {
  return sideMap[ []rune(s)[0] ]
}

func json2TVal(x interface{}) interface{} {
  var v interface{}
  switch vt := x.(type) {
  case string:
    if vt == "none" {
      v = thing.VT_NONE
    } else {
      v = thing.VT_UNLTD
    }
  case float64:
    v = vt
  }
  return v
}

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("load.go: " + fmtstr, args...))
}

type LoadFunc func([]interface{}) error

// loadRoom()
// [ ref, title, nav_ref_targets... ]
//
func loadRoom(data []interface{}) error {
  if len(data) < 2 {
    log(dtalog.ERR, "loadRoom(%q): argument slice not long enough", data)
    return fmt.Errorf("argument slice %q not long enough", data)
  }
  dat_strs := make([]string, 0, len(data))
  for _, x := range data {
    dat_strs = append(dat_strs, x.(string))
  }
  
  room.NewRoom(dat_strs[0], dat_strs[1], dat_strs[2:]...)
  return nil
}

// loadItem()
// [ ref, artAdjNoun, prep, plural, mass, bulk ]
//
func loadItem(x []interface{}) error {
  r          := x[0].(string)
  artAdjNoun := x[1].(string)
  prep       := x[2].(string)
  pl         := x[3].(bool)
  m := json2TVal(x[4])
  b := json2TVal(x[5])
  
  thing.NewItem(r, artAdjNoun, prep, pl, m, b)
  return nil
}

// loadItemContainer()
// [ref, artAdjNoun, prep, plural, mass, bulk, toggleable, open, 
// { "i": [mass_limit, bulk_limit], "o": [mass_limit, bulk_limit] ... } ]
//
func loadItemContainer(data []interface{}) error {
  loadItem(data[:6])
  nip := ref.Deref(data[0].(string))
  nicp := &thing.ItemContainer{
    Item: *(nip.(*thing.Item)),   // LOL, Jesus
    WillToggle: data[6].(bool),
    OpenState:  data[7].(bool),
    Sides: make(map[byte]*thing.ThingList),
  }
  
  sides := data[8].(map[string]interface{})
  
  for sid, mbl := range sides {
    s := str2side(sid)
    mbl := mbl.([]interface{})
    nicp.AddSide(s, json2TVal(mbl[0]), json2TVal(mbl[1]))
  }
  
  ref.Reregister(nicp)
  return nil
}

// loadDoorway()
// [ ref, artAdjNoun, prep, plural, mass, bulk, will_toggle ]
//
func loadDoorway(data []interface{}) error {
  loadItem(data[:6])
  nip := ref.Deref(data[0].(string))
  ndwyp := &door.Doorway{
    Item: *(nip.(*thing.Item)),
    WillToggle: data[6].(bool),
  }
  ref.Reregister(ndwyp)
  return nil
}

// loadDoor()
// [ dwy0, dwy1, isOpen ]
//
func loadDoor(data []interface{}) error {
  dwy0 := ref.Deref(data[0].(string)).(*door.Doorway) // again, Jesus
  dwy1 := ref.Deref(data[1].(string)).(*door.Doorway)
  is_open := data[2].(bool)
  door.Bind(dwy0, dwy1, is_open)
  return nil
}

// loadClothing()
// [ "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk, "slot" ]
//
func loadClothing(data []interface{}) error {
  loadItem(data[:6])
  nip := ref.Deref(data[0].(string)).(*thing.Item)
  slot := data[6].(string)
  thing.MakeClothing(nip, slot)
  return nil
}

// loadWornContainer()
// [ "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk,
//   "slot", will_toggle, is_open, mass_held, bulk_held ]
func loadWornContainer(data []interface{}) error {
  loadItem(data[:6])
  nip := ref.Deref(data[0].(string)).(*thing.Item)
  slot := data[6].(string)
  will_toggle := data[7].(bool)
  is_open     := data[8].(bool)
  mass_held   := json2TVal(data[9])
  bulk_held   := json2TVal(data[10])
  thing.MakeWornContainer(nip, slot, will_toggle, is_open, mass_held, bulk_held)
  return nil
}

// populate()
// [ ref, side, refs... ]
//
func populate(data []interface{}) error {
  r    := data[0].(string)
  side := sideMap[ []rune(data[1].(string))[0] ]
  cont := ref.Deref(r)
  
  var targ *thing.ThingList
  
  switch c := cont.(type) {
  case *room.Room:
    if side == room.CONTENTS {
      targ = c.Contents
    } else if side == room.SCENERY {
      targ = c.Scenery
    } else {
      log(dtalog.ERR, "populate(): unrecognized side identifier %q => %v for container type %T",
                      data[1], side, c)
      return fmt.Errorf("%q is not a recognizable side indentifier", data[1])
    }
  case thing.Container:
    targ = c.Side(side)
  }
  
  for _, t_ref := range data[2:] {
    targ.Add(ref.Deref(t_ref.(string)).(thing.Thing))
  }
  
  return nil
}

// loadMoodMessenger()
// [ min_secs, max_secs, [ room_refs... ], messages... ] 
//
func loadMoodMessenger(data []interface{}) error {
  min := data[0].(float64)
  max := data[1].(float64)
  raw_refs := data[2].([]interface{})
  refs := make([]string, 0, len(raw_refs))
  for _, r := range raw_refs {
    refs = append(refs, r.(string))
  }
  raw_msgs := data[3:]
  msgs := make([]string, 0, len(raw_msgs))
  for _, m := range raw_msgs {
    msgs = append(msgs, m.(string))
  }
  
  mood.NewMessenger(min, max, refs, msgs)
  return nil
}

// bindScript()
// [ "obj_ref", "verb", "script_tag" ]
//
func bindScript(data []interface{}) error {
  obj := ref.Deref(data[0].(string)).(thing.Thing)
  v   := data[1].(string)
  s   := data[2].(string)
  
  scripts.Bind(obj, v, s)
  return nil
}

func loadData(data []interface{}) error {
  datam := data[0].(map[string]interface{})
  thing.Data = make(map[string]map[string]interface{})
  for t_ref, t_dat := range datam {
    thing.Data[t_ref] = t_dat.(map[string]interface{})
  }
  return nil
}
  

var initialLoadMap = map[string]LoadFunc {
  "room":   loadRoom,
  "item":   loadItem,
  "itemc":  loadItemContainer,
  "dwy":    loadDoorway,
  "door":   loadDoor,
  "cloth":  loadClothing,
  "clothc": loadWornContainer,
  "pop":    populate,
  "mood":   loadMoodMessenger,
  "script": bindScript,
  "build":  build.Build,
  "data":   loadData,
}

var permanentLoadMap = map[string]LoadFunc {
  "room":   loadRoom,
  "dwy":    loadDoorway,
  "mood":   loadMoodMessenger,
  "script": bindScript,
}
var mutableLoadMap = map[string]LoadFunc {
  "item":   loadItem,
  "itemc":  loadItemContainer,
  "door":   loadDoor,
  "cloth":  loadClothing,
  "clothc": loadWornContainer,
  "pop":    populate,
  "data":   loadData,
}

type LoadType byte
const(  INIT  LoadType = 0
        PERM  LoadType = 1
        MUT   LoadType = 2
)

func LoadFeature(x []interface{}, mode LoadType) error {
  var lf LoadFunc
  var ok bool
  cmd := x[0].(string)
  switch mode {
  case INIT:
    lf, ok = initialLoadMap[cmd]
  case PERM:
    lf, ok = permanentLoadMap[cmd]
  case MUT:
    lf, ok = mutableLoadMap[cmd]
  }
  
  if ok {
    return lf(x[1:])
  } else {
    return nil
  }
}

func LoadFile(path string, mode LoadType) error {
  f, err := os.Open(path)
  if err != nil {
    log(dtalog.ERR, "LoadFile(%q): unable to open file", path)
    return fmt.Errorf("unable to open file %q", path)
  }
  defer f.Close()
  log(dtalog.DBG, "LoadFile(%q): file opened", path)
  
  dcdr := json.NewDecoder(f)

  for dcdr.More() {
    var x []interface{}
    err = dcdr.Decode(&x)
    if err != nil {
      log(dtalog.ERR, "LoadFile(%q): error in dcdr.Decode(): %s", path, err)
      return err
    }
    
    if x[0].(string) == "load" {
      LoadFile(filepath.Join(WorldDir, x[1].(string)), mode)
    } else {
      LoadFeature(x, mode)
    }
  }
  log(dtalog.DBG, "LoadFile(%q): done", path)
  return nil
}
