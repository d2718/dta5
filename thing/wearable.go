// wearable.go
//
// wearable dta5 items
//
// updated 2017-08-11
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
