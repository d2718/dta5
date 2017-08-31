// wearable.go
//
// wearable dta5 items
//
// updated 2017-08-12
//
package thing

import( "dta5/ref"; "dta5/save";
)

// Something Wearable returns the string representing the slot upon which
// it can be worn.
//
type Wearable interface {
  Slot() string
}

// A piece of Clothing is the most basic Thing that can be worn.
//
type Clothing struct {
  Item
  slot string
}

func (c Clothing) Slot() string {
  return c.slot
}

// Creates, ref.Register()s, and returns a new piece of Clothing.
//
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

// MakeClothing() takes an Item and makes it into a piece of Clothing that
// is worn in the supplied slot.
//
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

// The WornContainer implements both the Container and the Wearable
// interfaces. It represents a piece of clothing that can store stuff in
// it, like a backpack or a jacket with pockets.
//
type WornContainer struct {
  Item
  slot       string
  willToggle bool
  openState  bool
  contents   *ThingList
}

// Creates, ref.Register()s, and returns a new WornContainer.
//
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

// MakeWornContainer() turns a regular Item into a WornContainer with the
// extra supplied features.
//
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
