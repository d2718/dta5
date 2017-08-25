// thing.go
//
// dta5 Thing interface and methods
//
// A Thing is the most general classification of "item" -- something with 
// which one can interact -- in the game. Even PlayerChars implement the
// Thing interface. 
//
// A Thing has a Name (see dta5/name), a mass, and a bulk (sort of like 
// volume); these can be finite (for things you might pick up) or infinite
// (for things that should remain where they are) -- see the typed value,
// TVal, below.
//
// A Thing also implements the desc.Interface (see dta5/desc, although this
// may later get merged with ref.Interface), and has a pointer to where it
// is (see LocVec, below).
//
// updated 2017-08-15
//
package thing

import( "fmt";
        "dta5/desc"; "dta5/log"; "dta5/msg"; "dta5/name"; "dta5/ref";
        "dta5/save"; "dta5/util";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("thing: " + fmtstr, args...))
}

// The ValueType is for classifying TVals.
//
type ValueType byte
const ( VT_NONE   ValueType = iota  // no capacity or no mass/bulk
        VT_LTD                      // limited numerical capacity or mass/bulk
        VT_UNLTD                    // unlimited capacity or too much mass/bulk
                                    //    to be held, carried, or placed
                                    //    in/on/behind anything
)

// TVal represents a "typed value", that can either be
//  * none (VT_NONE)
//  * some arbitrary limit (VT_LTD)
//  * unlimited (VT_UNLTD)
//
type TVal struct {
  VT    ValueType
  Value float32       // should only matter if TVal.VT has a value of VT_LTD
}

var NONE  TVal = TVal{ VT: VT_NONE,  Value: 0.0 } // I put these here for convenience,
var INFTY TVal = TVal{ VT: VT_UNLTD, Value: 0.0 } // but I'm not sure I've ever used them.

// (TVal) Save() is used in saving to turn a TVal into something that can be
// easily human-read and also interpreted by the dta5/load package.
//
func (tv TVal) Save() interface{} {
  switch tv.VT {
  case VT_NONE:
    return "none"
  case VT_LTD:
    return tv.Value
  default:
    return "x"
  }
}

// A LocVec, (Location Vector) describes where a Thing is. It has a pointer to
// the containing object, and a byte where in the object it is stored. For
// containers that are other Things, see the constants below; for containers
// that are room.Rooms, see the dta5/room package.
//
type LocVec struct {
  Place ref.Interface
  Side  byte
}

const(  IN     byte = 0
        ON     byte = 1
        BEHIND byte = 2
        UNDER  byte = 3
)

// Turns the Side member of a LocVec into a descriptive string.
//
func sideStr(s byte) string {
  switch s {
  case IN:
    return "in"
  case ON:
    return "on"
  case BEHIND:
    return "behind"
  case UNDER:
    return "under"
  default:
    return ""
  }
}

func SideStr(s byte) string {
  return sideStr(s)
}

func (v LocVec) String() string {
  if t, ok := v.Place.(Thing); ok {
    return fmt.Sprintf("%s %s", sideStr(v.Side), t.Normal(0))
  } else {
    return ""
  }
}

type Thing interface {
  ref.Interface
  name.Name
  Mass() TVal
  Bulk() TVal
  desc.Interface
  Loc()  LocVec
  SetLoc(LocVec)
  Deliver(*msg.Message)
  Save(save.Saver)
}

// A ThingList represents a place where Things can be, such as inside a
// container or on the ground in a room.Room. It can have limited containment
// capacity.
//
type ThingList struct {
  Things []Thing
  MassLimit TVal
  BulkLimit TVal
  LocVec
}

// NewThingList()
//   * mass, bulk are either VT_NONE, VT_UNLTD, or a floating-point value;
//     they will get converted to the proper TVal.
//   * r is the referent that holds the ThingList
//   * s is the side of that referent where the ThingList is
//     (one of IN, ON, BEHIND, UNDER, or see dta5/room)
//
func NewThingList(mass, bulk interface{}, r ref.Interface, s byte) *ThingList {
  var new_mass, new_bulk TVal
  switch m := mass.(type) {
  case ValueType:
    new_mass = TVal{ VT: m, Value: 0.0 }
  case float32:
    new_mass = TVal{ VT: VT_LTD, Value: m }
  case float64:
    new_mass = TVal{ VT: VT_LTD, Value: float32(m) }
  default:
    log(dtalog.MSG, "NewThingList(): bad type for mass (%T)", mass)
    new_mass = TVal{ VT: VT_NONE, Value: 0.0 }
  }
  switch b := bulk.(type) {
  case ValueType:
    new_bulk = TVal{ VT: b, Value: 0.0 }
  case float32:
    new_bulk = TVal{ VT: VT_LTD, Value: b }
  case float64:
    new_bulk = TVal{ VT: VT_LTD, Value: float32(b) }
  default:
    log(dtalog.MSG, "NewThingList(): bad type for bulk (%T)", bulk)
    new_bulk = TVal{ VT: VT_NONE, Value: 0.0 }
  }
  return &ThingList{ Things: make([]Thing, 0, 0),
                     MassLimit: new_mass, BulkLimit: new_bulk,
                     LocVec: LocVec{ Place: r, Side: s}}
}

// Return the total amount of mass contained in the ThingList.
//
func (tl ThingList) TotalMass() TVal {
  var tot float32 = 0.0
  for _, t := range tl.Things {
    m := t.Mass()
    switch m.VT {
    case VT_UNLTD:
      return TVal{ VT: VT_UNLTD, Value: tot, }
    case VT_LTD:
      tot += m.Value
    }
  }
  return TVal{ VT: VT_LTD, Value: tot, }
}

// Return the total bulk of thing inside the ThingList.
//
func (tl ThingList) TotalBulk() TVal {
  var tot float32 = 0.0
  for _, t := range tl.Things {
    b := t.Bulk()
    switch b.VT {
    case VT_UNLTD:
      return TVal{ VT: VT_UNLTD, Value: tot, }
    case VT_LTD:
      tot += b.Value
    }
  }
  return TVal{ VT: VT_LTD, Value: tot, }
}

// Adds t to the ThingList, updating its location. Checking whether t will
// fit is the calling function's responsibility.
//
func (tl *ThingList) Add(t Thing) {
  log(dtalog.DBG, "(*ThingList [%q, %d]) Add(%q) called",
                    tl.LocVec.Place.Ref(), tl.LocVec.Side, t.Ref())
  tl.Things = append(tl.Things, t)
  t.SetLoc(tl.LocVec)
}

// Removes t from the ThingList, setting its location to be nothing. You
// probably shouldn't call this unless you're sure t is in there.
//
func (tl *ThingList) Remove(t Thing) {
  log(dtalog.DBG, "(*ThingList [%q, %d]) Remove(%q) called",
                  tl.LocVec.Place.Ref(), tl.LocVec.Side, t.Ref())
  if len(tl.Things) < 1 {
    return
  }
  new_slice := make([]Thing, 0, len(tl.Things)-1)
  for _, x := range tl.Things {
    if x != t {
      new_slice = append(new_slice, x)
    }
  }
  tl.Things = new_slice
  log(dtalog.DBG, "Remove(): new Things: %s", tl.EnglishList())
  t.SetLoc(LocVec{ Place: nil, Side: 0 })
}

// Return the ordth Thing whose name matches the given tokens, or nil,
// with ord reduced by the number of Things that did match.
//
func (tl ThingList) Find(toks []string, ord int) (Thing, int) {
  var remain int = ord
  for _, t := range tl.Things {
    if t.Match(toks) {
      if remain == 0 {
        return t, 0
      } else {
        remain--
      }
    }
  }
  return nil, remain
}

// Return true if t is in the ThingList.
//
func (tl ThingList) Contains(t Thing) bool {
  for _, x := range tl.Things {
    if t == x {
      return true
    }
  }
  return false
}

// Return true if adding t to the ThingList will not cause its mass capacity
// to be exceeded.
//
func (tl ThingList) WillFitMass(t Thing) bool {
  switch tl.MassLimit.VT {
  case VT_UNLTD:
    return true
  case VT_NONE:
    return false
  default:
    nm := t.Mass()
    if nm.VT == VT_UNLTD {
      return false
    } else if nm.VT == VT_NONE {
      return true
    }
    cur_amt := tl.TotalMass()
    if cur_amt.VT == VT_UNLTD {
      return false
    }
    if cur_amt.Value + nm.Value > tl.MassLimit.Value {
      return false
    }
  }
  return true
}

// Return true if adding t to the ThingList will not cause its bulk capacity
// to be exceeded.
//
func (tl ThingList) WillFitBulk(t Thing) bool {
  switch tl.BulkLimit.VT {
  case VT_UNLTD:
    return true
  case VT_NONE:
    return false
  default:
    nb := t.Bulk()
    if nb.VT == VT_UNLTD {
      return false
    } else if nb.VT == VT_NONE {
      return true
    }
    cur_amt := tl.TotalBulk()
    if cur_amt.VT == VT_UNLTD {
      return false
    }
    if cur_amt.Value + nb.Value > tl.BulkLimit.Value {
      return false
    }
  }
  return true
}

// Deliver the given message to all Things in the ThingList.
//
func (tl ThingList) Deliver(m *msg.Message) {
  for _, t := range tl.Things {
    t.Deliver(m)
  }
}

// Return a string containing the names of all contained things as an
// appropriately comma'd list, or "nothing" if empty.
//
func (tl ThingList) EnglishList() string {
  if len(tl.Things) == 0 {
    return "nothing"
  } else {
    namz := make([]string, 0, len(tl.Things))
    for _, t := range tl.Things {
      namz = append(namz, t.Normal(0))
    }
    return util.EnglishList(namz)
  }
}

// Calls the given function on each Thing in the ThingList, and recursively
// on all objects in contained containers.
//
func (tl ThingList) Walk(f func(Thing)) {
  for _, t := range tl.Things {
    f(t)
    if ot, ok := t.(Container); ok {
      for _, s := range []byte{IN, ON, BEHIND, UNDER} {
        if tk := ot.Side(s); tk != nil {
          tk.Walk(f)
        }
      }
    }
  }
}

// Most basic struct that implements the Thing interface.
//
type Item struct {
  ref string
  name.NormalName
  descPage *string
  mass TVal
  bulk TVal
  where LocVec
}

// Create and ref.Register() a new Item with the given parameters:
//  * nref is the Item's reference string
//  * artAdjNoun, prep, and plural specify the Name (see dta5/name)
//  * mass, bulk can be VT_NONE, VT_UNLTD, or a floating-point value
//        which will be converted to the appropriate TVal
//
func NewItem(nref, artAdjNoun, prep string, plural bool,
             mass, bulk interface{}) *Item {
               
  nam := name.NewNormal(artAdjNoun, prep, plural)
  var mtv, btv TVal
  switch m := mass.(type) {
  case ValueType:
    mtv = TVal{ VT: m, Value: 0.0, }
  case float32:
    mtv = TVal{ VT: VT_LTD, Value: m, }
  case float64:
    mtv = TVal{ VT: VT_LTD, Value: float32(m), }
  default:
    log(dtalog.MSG, "NewItem(): bad type (%T) for mass")
    mtv = TVal{ VT: VT_NONE, Value: 0.0, }
  }
  switch b := bulk.(type) {
  case ValueType:
    btv = TVal{ VT: b, Value: 0.0, }
  case float32:
    btv = TVal{ VT: VT_LTD, Value: b, }
  case float64:
    btv = TVal{ VT: VT_LTD, Value: float32(b), }
  default:
    log(dtalog.MSG, "NewItem(): bad type (%T) for bulk")
    btv = TVal{ VT: VT_NONE, Value: 0.0, }
  }
  
  rval := Item{ ref: nref,
                NormalName: *nam,
                descPage: nil,
                mass: mtv,
                bulk: btv,
              }
  err := ref.Register(&rval)
  if err != nil {
    return nil
  } else {
    return &rval
  }
}

// Item implements ref.Interface
//
func (i Item) Ref() string { return i.ref }
func (i Item) Data(key string) interface{} { return ref.GetData(i, key) }
func (i Item) SetData(key string, val interface{}) { ref.SetData(i, key, val) }

func (i Item) Mass() TVal { return i.mass }
func (i Item) Bulk() TVal { return i.bulk }

// Item implements desc.Interface
func (i *Item) SetDescPage(pagep *string) { i.descPage = pagep }
func (i Item) Desc() string {
  if i.descPage == nil {
    return "You notice nothing special."
  }
  return desc.GetDesc(*(i.descPage), i.ref)
}

// Return the Item's location.
func (i Item) Loc() LocVec { return i.where }
// Set the Item's location.
func (ip *Item) SetLoc(lv LocVec) {ip.where = lv }

// Item's don't care about messages, so just do nothing.
//
func (i Item) Deliver(m *msg.Message) { return }

// Saves the state of the Item. See dta5/save and dta5/load for how this works.
//
func (i Item) Save(s save.Saver) {
  var cont []interface{} = make([]interface{}, 0, 7)
  cont = append(cont, "item")
  cont = append(cont, i.ref)
  cont = append(cont, i.NormalName.ToSaveString())
  cont = append(cont, i.NormalName.PrepPhrase)
  cont = append(cont, false)
  cont = append(cont, i.mass.Save())
  cont = append(cont, i.bulk.Save())
  s.Encode(cont)
}
