// wearable.go
//
// wearable dta5 items
//
// updated 2017-08-12
//
package thing

import( "dta5/ref"; "dta5/save";
)

type Wearable interface {
  Slot() string
}

type Clothing struct {
  Item
  slot string
}

func (c Clothing) Slot() string {
  return c.slot
}

func NewClothing(nref, artAdjNoun, prep string, plural bool,
                 mass, bulk interface{}, wornSlot string) *Clothing {
  nip := NewItem(nref, artAdjNoun, prep, plural, mass, bulk)
  nc := Clothing{
    Item: *nip,
    slot: wornSlot,
  }
  ref.Register(&nc)
  return &nc
}

func MakeClothing(ip *Item, wornSlot string) *Clothing {
  nc := Clothing{
    Item: *ip,
    slot: wornSlot,
  }
  ref.Reregister(&nc)
  return &nc
}

func (c Clothing) Save(s save.Saver) {
  var data []interface{} = []interface{} {
    "cloth", c.Ref(), c.NormalName.ToSaveString(), c.NormalName.PrepPhrase,
    false, c.mass.Save(), c.bulk.Save(), c.slot, }
  s.Encode(data)
}

type WornContainer struct {
  Item
  slot       string
  willToggle bool
  openState  bool
  contents   *ThingList
}

func NewWornContainer(nref, artAdjNoun, prep string, plural bool,
                      mass, bulk interface{}, wornSlot string,
                      toggleable, isOpen bool, massHeld,
                      bulkHeld interface{}) *WornContainer {
  nip := NewItem(nref, artAdjNoun, prep, plural, mass, bulk)
  nwc := WornContainer{
    Item:       *nip,
    slot:       wornSlot,
    willToggle: toggleable,
    openState:  isOpen,
    contents:   nil,
  }
  
  ref.Reregister(&nwc)
  nwc.contents = NewThingList(massHeld, bulkHeld, &nwc, IN)
  return &nwc
}

func (w WornContainer) IsToggleable() bool { return w.willToggle }
func (w WornContainer) IsOpen() bool { return w.openState }
func (wp *WornContainer) SetToggleable(b bool) { wp.willToggle = b }
func (wp *WornContainer) SetOpen(b bool) { wp.openState = b }
func (w WornContainer) Slot() string { return w.slot }

func (wp *WornContainer) Side(s byte) *ThingList {
  if s == IN {
    return wp.contents
  } else {
    return nil
  }
}

func MakeWornContainer(ip *Item, wornSlot string, toggleable, isOpen bool,
                        massHeld, bulkHeld interface{}) *WornContainer {
  nwc := WornContainer{
    Item:       *ip,
    slot:       wornSlot,
    willToggle: toggleable,
    openState:  isOpen,
    contents:   nil,
  }
  
  ref.Reregister(&nwc)
  nwc.contents = NewThingList(massHeld, bulkHeld, &nwc, IN)
  return &nwc
}

func (w WornContainer) Save(s save.Saver) {
  var data []interface{} = []interface{} {
    "clothc", w.Ref(), w.NormalName.ToSaveString(), w.NormalName.PrepPhrase,
    false, w.mass.Save(), w.bulk.Save(), w.slot, w.willToggle, w.openState,
    w.contents.MassLimit.Save(), w.contents.BulkLimit.Save(), }
  s.Encode(data)
  
  if len(w.contents.Things) > 0 {
    var pop_data []interface{} = []interface{} { "pop", w.Ref(), sideStr(IN), }
    for _, t := range w.contents.Things {
      t.Save(s)
      pop_data = append(pop_data, t.Ref())
    }
    s.Encode(pop_data)
  }
}
