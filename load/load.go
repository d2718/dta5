// load.go
//
// dta5 loading and world-building module
//
// updated 2017-08-12
//
// Worlds in dta5 are specified (and saved!) as JSON files; each JSON file
// contains a series of lists. The first element in each list is a string
// designating which type of object is to be created (e.g., "room" is to
// create a room.Room) or action to be taken (e.g., "load" to load the
// contents of another file); the following elements are parameters for the
// associated object or action.
//
// The various formats are listed here as a shorthand; see the individual
// loadXXX() functions below for more details.
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
// ["dwy", "ref", "artAdjNoun", "prepPhrase", plural, mass, bulk, WillToggle ]
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

// The various loadXXX() functions use this to help translate human-readable
// side designation strings into the appropriate byte values.
//
var sideMap map[rune]byte = map[rune]byte {
  's':  room.SCENERY,
  'c':  room.CONTENTS,
  'o':  thing.ON,
  'i':  thing.IN,
  'b':  thing.BEHIND,
  'u':  thing.UNDER,
}

// This converts a human-readable string to the appropriate byte representing
// a location's side.
//
func str2side(s string) byte {
  return sideMap[ []rune(s)[0] ]
}

// Converts a JSON value (string of float) into a thing.TVal.
//
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
  dtalog.Log(lvl, fmt.Sprintf("load: " + fmtstr, args...))
}

// All of the various loadXXX() functions that operate on the individual
// JSON lists are of this type.
//
type LoadFunc func([]interface{}) error

// loadRoom()
// [ ref, title, nav_ref_targets... ]
//
// Creates a room.Room
//   * ref   string: the Room's reference string
//   * title string: the name of the Room
//   * nav_ref_targets string...: reference strings of the targets of moving
//        in the cardinal directions from this room (either other Rooms, or
//        door.Doorways; in the order specified in the const section of
//        room/room.go
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
// Creates a thing.Item
//   * ref string: the Item's reference string
//   * artAdjNoun, prep string, plural bool: the Item's name (see dta5/name)
//   * mass, bulk interface{}: to be read by json2TVal()
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
//       { "i": [mass_limit, bulk_limit], "o": [mass_limit, bulk_limit] ... } ]
//
// Creates a thing.ItemContainer
//   * ref, artAdjNoun, prep, plural, mass, bulk: see loadItem() above
//   * toggleable bool: whether the container can be opened/closed
//   * open bool: whether it begins in the open state
//   * the final element is a map that specifies the limits of the various
//         sides of the container; keys should be from sideMap (above), and
//         [mass, bulk] tuples should have values readable by json2TVal
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
// Creates a door.Doorway
//   * ref, artAdjNoun, prep, plural, mass, bulk: see loadItem() above
//   * will_toggle bool: whether the doorway can be opened/closed
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
// Creates a door.Door (that is, a binding between two door.Doorways to
// create a passage between Rooms)
//   * dwy0, dwy1 string: refs of the doorways to bind
//   * isOpen bool: whether the Door starts in the open state
//
func loadDoor(data []interface{}) error {
  dwy0 := ref.Deref(data[0].(string)).(*door.Doorway) // again, Jesus
  dwy1 := ref.Deref(data[1].(string)).(*door.Doorway)
  is_open := data[2].(bool)
  door.Bind(dwy0, dwy1, is_open)
  return nil
}

// loadClothing()
// [ ref, artAdjNoun, prepPhrase, plural, mass, bulk, slot ]
//
// Creates a thing.Clothing
//   * ref, artAdjNoun, prepPhrase, plural, mass, bulk: see loadItem() above
//   * slot string: the name of the clothing slot this occupies (see
//         thing/wearable.go and the dta5/body package
//
func loadClothing(data []interface{}) error {
  loadItem(data[:6])
  nip := ref.Deref(data[0].(string)).(*thing.Item)
  slot := data[6].(string)
  thing.MakeClothing(nip, slot)
  return nil
}

// loadWornContainer()
// [ ref, artAdjNoun, prepPhrase, plural, mass, bulk, slot,
//        will_toggle, is_open, mass_held, bulk_held ]
//
// Creates a thing.WornContainer
//   * ref, artAdjNoun, prepPhrase, plural, mass, bulk, slot: see loadClothing()
//   * will_toggle bool: whether the container is opeable/closable
//   * is_open bool: whether the container begins in the open state
//   * mass_held, bulk_held: to be read by json2Tval()
//
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
// Puts thing.Things in a room.Room or a thing.Container.
//   * ref string: the reference string of the Room/Container to load
//   * side string: should start with an appropriate letter (see sideMap above)
//   * refs string...: reference strings of the Things to get loaded
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
// Create a mood.MoodMessenger
//   * min_secs, max_secs float: limits of how often messages are delivered
//   * room_refs [ string ... ]: refs of Rooms where messages are delivered
//   * messages string...: message strings
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
// [ obj_ref, verb, script_tag ]
//
// Bind a script (see dta5/scripts) to a thing.Thing
//   * ref string: reference of Thing to bind
//   * verb string string: verb that will trigger the script
//   * script_tag string: key in scripts.Scripts that maps to the function
//         to bind
//
func bindScript(data []interface{}) error {
  obj := ref.Deref(data[0].(string)).(thing.Thing)
  v   := data[1].(string)
  s   := data[2].(string)
  
  scripts.Bind(obj, v, s)
  return nil
}

// loadData()
// [ stuff ]
//
// Loads ref.Data; you probably don't ever need to use this explicitly; it'll
// get written and read during the saving/loading of game states.
//
func loadData(data []interface{}) error {
  datam := data[0].(map[string]interface{})
  ref.Data = make(map[string]map[string]interface{})
  for t_ref, t_dat := range datam {
    ref.Data[t_ref] = t_dat.(map[string]interface{})
  }
  return nil
}

// These three maps specify which type of loading should be done under
// different circumstances. Obviously, when the game is first started, all of
// the world data should be loaded. Only "non-permanent" objects and states
// are saved when the game is saved, though, so when the game is loaded,
// "permanent" objects should be read from the "world" file, and
// "non-permanent" objects and state from the save file.

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

// These values are used to specify what type of loading situation is
// taking place when calling the LoadFeature() and LoadFile() functions
// below:
//
// INIT: Initial loading of the game world upon game launch.
// PERM: Reading of game world files for loading of "permanent" game world
//       features upon the loading of a saved game state.
// MUT : Reading of saved game file(s) upon loading of a saved game state.
//
type LoadType byte
const(  INIT  LoadType = 0
        PERM  LoadType = 1
        MUT   LoadType = 2
)

// LoadFeature() is called for every item (that is, JSON list specifying an
// element of the game world) in a loaded file. Given the conditions (that is,
// the current mode), the item is either ignored, or the appropriate loadXxx()
// function is called.
//
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

// LoadFile() reads sequentially the JSON lists in a given file specifying
// world data and invokes the appropriate functions (by calling LoadFeature())
// or reads the appropriate linked files (by calling itself).
//
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
