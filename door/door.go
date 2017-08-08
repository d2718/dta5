// door.go
//
// dta5 doors and doorways
//
// updated 2017-08-04
//
package door

import( "fmt";
        "dta5/log"; "dta5/ref"; "dta5/thing"; "dta5/save"
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("door: " + fmtstr, args...))
}

type Doorway struct {
  thing.Item
  WillToggle bool
  binder *Door
}

type Door struct {
  Side0  *Doorway
  Side1  *Doorway
  IsOpen bool
}

var Doors []*Door = make([]*Door, 0, 0)

func Reset() {
  Doors = make([]*Door, 0, 0)
}

func New(nref, artAdjNoun, prep string, mass, bulk interface{},
         toggleable bool) *Doorway {
  ni := thing.NewItem(nref, artAdjNoun, prep, false, mass, bulk)
  ndwy := Doorway{ Item: *ni, WillToggle: toggleable, binder: nil, }
  ref.Register(&ndwy)
  return &ndwy
}

func Bind(dwy0, dwy1 *Doorway, isOpen bool) {
  log(dtalog.DBG, "Bind(%q, %q, %v) called", dwy0.Ref(), dwy1.Ref(), isOpen)
  
  if dwy0.binder != nil {
    log(dtalog.WRN, "Bind(): rebinding %q", dwy0.Ref())
  }
  if dwy1.binder != nil {
    log(dtalog.WRN, "Bind(): rebinding %q", dwy1.Ref())
  }
  
  nd := Door{ Side0: dwy0, Side1: dwy1, IsOpen: isOpen, }
  Doors = append(Doors, &nd)
  dwy0.binder = &nd
  dwy1.binder = &nd
}

func (dwy Doorway) IsToggleable() bool {
  return dwy.WillToggle
}
func (dwyp *Doorway) SetToggleable(isToggleable bool) {
  dwyp.WillToggle = isToggleable
}
func (dwy Doorway) IsOpen() bool {
  return dwy.binder.IsOpen
}
func (dwyp *Doorway) SetOpen(isOpen bool) {
  dwyp.binder.IsOpen = isOpen
}

func (dwyp *Doorway) Other() *Doorway {
  if dwyp == dwyp.binder.Side0 {
    return dwyp.binder.Side1
  } else {
    return dwyp.binder.Side0
  }
}

func (dwy Doorway) Save(s save.Saver) {
  var cont []interface{} = make([]interface{}, 0, 8)
  cont = append(cont, "dwy")
  cont = append(cont, dwy.Ref())
  cont = append(cont, dwy.NormalName.ToSaveString())
  cont = append(cont, dwy.NormalName.PrepPhrase)
  cont = append(cont, false)
  cont = append(cont, dwy.Mass().Save())
  cont = append(cont, dwy.Bulk().Save())
  cont = append(cont, dwy.WillToggle)
  s.Encode(cont)
}

func (d Door) Save(s save.Saver) {
  var cont []interface{} = make([]interface{}, 0, 4)
  cont = append(cont, "door")
  cont = append(cont, d.Side0.Ref())
  cont = append(cont, d.Side1.Ref())
  cont = append(cont, d.IsOpen)
  s.Encode(cont)
}
