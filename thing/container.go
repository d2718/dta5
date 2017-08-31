// container.go
//
// dta5 Container interface and basic container implementation
//
// updated 2017-08-30
//
// A Container is something that can store other Things (generally in
// some combination of in, on, behind, and under it).
//
package thing

import(
        "dta5/name"; "dta5/ref"; "dta5/save";
)

// Things that can be opened and closed should implement this interface.
//
type Openable interface {
  IsToggleable() bool
  SetToggleable(bool)
  IsOpen() bool
  SetOpen(bool)
}

// Any Thing that can store other Things should implement this interface.
// For suggested values of the argument to Side(), the constants IN, ON,
// BEHIND, and UNDER are defined in the file thing.go.
//
type Container interface {
  name.Name
  Side(byte) *ThingList
  Openable
}

// The reference implementation of thing.Container, it has a map of
// *ThingLists with keys IN, ON, BEHIND, and UNDER (any of which might map
// to nil if it can't store objects there).
//
type ItemContainer struct {
  Item
  WillToggle bool
  OpenState bool
  Sides map[byte]*ThingList
}

// Creates, ref.Register()s, and returns a new *ItemContainer. If you actually
// want to store stuff in it, you need to use the AddSide() method.
//
func NewItemContainer(nref, artAdjNoun, prep string, plural bool,
                      descFile *string, mass, bulk interface{},
                      toggleable, open bool) *ItemContainer {
  ni := NewItem(nref, artAdjNoun, prep, plural, mass, bulk)
  nic := ItemContainer{
    Item: *ni,
    WillToggle: toggleable,
    OpenState:  open,
    Sides: make(map[byte]*ThingList),
  }
  
  ref.Reregister(&nic)
  return &nic
}

// Makes an ItemContainer actually able to store stuff. For example,
//
//  c := thing.NewItemContainer("r0-t0", "a/an massive marble altar", "",
//                              false, nil, thing.VT_UNLTD, thing.VT_UNLTD,
//                              false, false)
//  c.AddSide(thing.BEHIND, thing.VT_UNLTD, 2000)
//  c.AddSide(thing.ON, 10000, 2000)
//
// will create an altar that people can put stuff on and behind.
//
func (ic *ItemContainer) AddSide(s byte, mass, bulk interface{}) {
  ntl := NewThingList(mass, bulk, ic, s)
  ic.Sides[s] = ntl
}

// Return the *ThingList that represents the inventory on the given side
// of the ItemContainer.
//
func (ic *ItemContainer) Side(s byte) *ThingList {
  tl, ok := ic.Sides[s]
  if ok {
    return tl
  } else {
    return nil
  }
}

func (ic ItemContainer) IsToggleable() bool {
  return ic.WillToggle
}
func (icp *ItemContainer) SetToggleable(isToggleable bool) {
  icp.WillToggle = isToggleable
}
func (ic ItemContainer) IsOpen() bool {
  return ic.OpenState
}
func (icp* ItemContainer) SetOpen(isOpen bool) {
  icp.OpenState = isOpen
}

func (ic ItemContainer) Save(s save.Saver) {
  var cont []interface{} = make([]interface{}, 0, 10)
  cont = append(cont, "itemc")
  cont = append(cont, ic.ref)
  cont = append(cont, ic.NormalName.ToSaveString())
  cont = append(cont, ic.NormalName.PrepPhrase)
  cont = append(cont, false)
  cont = append(cont, ic.mass.Save())
  cont = append(cont, ic.bulk.Save())
  cont = append(cont, ic.WillToggle)
  cont = append(cont, ic.OpenState)
  side_info := make(map[string]interface{})
  for sid, tl := range ic.Sides {
    mblim := []interface{}{ tl.MassLimit.Save(), tl.BulkLimit.Save(), }
    side_info[sideStr(sid)] = mblim
  }
  cont = append(cont, side_info)
  s.Encode(cont)
  
  for sid, tl := range ic.Sides {
    if len(tl.Things) > 0 {
      for _, t := range tl.Things {
        t.Save(s)
      }
      pop_cont := make([]interface{}, 0, len(tl.Things) + 3)
      pop_cont = append(pop_cont, "pop")
      pop_cont = append(pop_cont, ic.Ref())
      pop_cont = append(pop_cont, sideStr(sid))
      for _, t := range tl.Things {
        pop_cont = append(pop_cont, t.Ref())
      }
      s.Encode(pop_cont)
    }
  }
}
