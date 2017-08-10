// thing.go
//
// dta5 Thing class and methods
//
// updated 2017-08-08
//
package thing

import( "fmt";
        "dta5/desc"; "dta5/log"; "dta5/msg"; "dta5/name"; "dta5/ref";
        "dta5/save"; "dta5/util";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("thing.go: " + fmtstr, args...))
}

type ValueType byte
const ( VT_NONE   ValueType = iota
        VT_LTD
        VT_UNLTD
)

// TVal represents a "typed value", that can either be
//  * none (VT_NONE)
//  * some arbitrary limit (VT_LTD)
//  * unlimited (VT_UNLTD)
//
type TVal struct {
  VT    ValueType
  Value float32
}

var NONE  TVal = TVal{ VT: VT_NONE,  Value: 0.0 }
var INFTY TVal = TVal{ VT: VT_UNLTD, Value: 0.0 }

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

type LocVec struct {
  Place ref.Interface
  Side  byte
}

const(  IN     byte = 0
        ON     byte = 1
        BEHIND byte = 2
        UNDER  byte = 3
)

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

var Data = make(map[string]map[string]interface{})

func SaveData(s save.Saver) {
  x := []interface{}{"data", Data, }
  s.Encode(x)
}

type Thing interface {
  Ref() string
  name.Name
  Mass() TVal
  Bulk() TVal
  desc.Interface
  Loc()  LocVec
  SetLoc(LocVec)
  Data(string) interface{}
  SetData(string, interface{})
  Deliver(*msg.Message)
  Save(save.Saver)
}

type ThingList struct {
  Things []Thing
  MassLimit TVal
  BulkLimit TVal
  LocVec
}

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

func (tl *ThingList) Add(t Thing) {
  log(dtalog.DBG, "(*ThingList [%q, %d]) Add(%q) called",
                    tl.LocVec.Place.Ref(), tl.LocVec.Side, t.Ref())
  tl.Things = append(tl.Things, t)
  t.SetLoc(tl.LocVec)
}

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

func (tl ThingList) Contains(t Thing) bool {
  for _, x := range tl.Things {
    if t == x {
      return true
    }
  }
  return false
}

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

func (tl ThingList) Deliver(m *msg.Message) {
  for _, t := range tl.Things {
    t.Deliver(m)
  }
}

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

type Item struct {
  ref string
  name.NormalName
  descPage *string
  mass TVal
  bulk TVal
  where LocVec
}

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

func (i Item) Ref() string {
  return i.ref
}

func (i Item) Mass() TVal {
  return i.mass
}
func (i Item) Bulk() TVal {
  return i.bulk
}

func (i *Item) SetDescPage(pagep *string) {
  i.descPage = pagep
}

func (i Item) Desc() string {
  if i.descPage == nil {
    return "You notice nothing special."
  }
  
  return desc.GetDesc(*(i.descPage), i.ref)
}

func (i Item) Loc() LocVec {
  return i.where
}
func (ip *Item) SetLoc(lv LocVec) {
  ip.where = lv
}

func (i Item) Data(tok string) interface{} {
  if Data[i.ref] == nil {
    return nil
  } else {
    return Data[i.ref][tok]
  }
}

func (i Item) SetData(tok string, val interface{}) {
  if Data[i.ref] == nil {
    Data[i.ref] = make(map[string]interface{})
  }
  Data[i.ref][tok] = val
}

func (i Item) Deliver(m *msg.Message) { return }

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
