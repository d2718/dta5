// container.go
//
// dta5 container interface and basic container implementation
//
// updated 2017-08-11
//
package thing

import(
        "dta5/name"; "dta5/ref"; "dta5/save";
)

type Openable interface {
  IsToggleable() bool
  SetToggleable(bool)
  IsOpen() bool
  SetOpen(bool)
}

type Container interface {
  name.Name
  Side(byte) *ThingList
  Openable
}

type ItemContainer struct {
  Item
  WillToggle bool
  OpenState bool
  Sides map[byte]*ThingList
}

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

func (ic *ItemContainer) AddSide(s byte, mass, bulk interface{}) {
  ntl := NewThingList(mass, bulk, ic, s)
  ic.Sides[s] = ntl
}

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
