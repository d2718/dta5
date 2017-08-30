// door.go
//
// dta5 doors and doorways
//
// updated 2017-08-30
//
// Doors represent portals between room.Rooms. Rooms can be connected
// directly via the cardinal directions, but if a richer passage between
// them is needed (a named object to go through, something that can be
// closed or locked), the Door comes into play.
//
// The Doorway (implementing thing.Thing) is a physical object to be placed
// in a Room's contents or scenery; two Doorways (generally in different
// Rooms) are bound together in a Door, so that entering one will cause
// a creature to exit the other, and opening/closing one will open/close
// the other.
//
package door

import( "fmt";
        "dta5/log"; "dta5/ref"; "dta5/thing"; "dta5/save"
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("door: " + fmtstr, args...))
}

// A Doorway represents a Door's physical presence in a Room; think of it as
// one side of a door. To link two Rooms, place a Doorway in each, and Bind()
// them together into a Door object.
//
// *Doorway implements the thing.Openable interface.
//
type Doorway struct {
  thing.Item
  WillToggle bool
  binder *Door
}

// A Door connects two Doorways together.
//
type Door struct {
  Side0  *Doorway
  Side1  *Doorway
  IsOpen bool
}

// Door objects never need to be placed anywhere in the game world; they don't
// have ref strings associated with them, but they do need somewhere to
// live, and that place is here.
//
var Doors []*Door = make([]*Door, 0, 0)

// Reset() should be called in preparation for populating the game world,
// either on initial load or when loading a saved state.
//
func Reset() {
  Doors = make([]*Door, 0, 0)
}

// Creates, ref.Register()s, and returns a *Doorway.
//
func New(nref, artAdjNoun, prep string, mass, bulk interface{},
         toggleable bool) *Doorway {
           
  ni := thing.NewItem(nref, artAdjNoun, prep, false, mass, bulk)
  ndwy := Doorway{ Item: *ni, WillToggle: toggleable, binder: nil, }
  ref.Register(&ndwy)
  return &ndwy
}

// Binds two Doorways together in a Door object, and stashes that Door in
// the Doors slice.
//
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

// Other() returns a pointer to the other half of the Doorway's Door.
//
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
