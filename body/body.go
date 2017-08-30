// body.go
//
// dta5 Body struct and methods
// Bodies are for Things (like PlayerChars) that can wear clothes and
// participate in combat.
//
// updated 2017-08-30
//
package body

import( "fmt";
        "dta5/log"; "dta5/thing"
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("body: " + fmtstr, args...))
}

// As of now, the body.Interface is mainly concerned with where thing.Things
// can be held and where thing.Wearables can be worn. Each thing.Wearable
// has a "slot" (a string uniquely representing a body part) where it can be
// worn. If a body.Interface has that slot, it can wear that thing.Wearable.
//
// For example, a human has two "shoulder" slots, so it could wear two
// separate things which could be slung over the shoulder; a snake, on the
// other hand, could not, although depending on its size, it _might_ be able
// to wear one or more things that occupy the "neck" slot.
//
// A body.Interface also has zero or more slots where thing.Things can be
// held. For a human, these two slots are "right_hand" and "left_hand".
//
// See the BasicBody type for an example of this implementation.
//
type Interface interface {
  WornSlots(string) (byte, bool)
  WornSlotKeys() []string
  WornSlotName(string) string
  HeldIn(string) (thing.Thing, bool)
  HeldSlotKeys() []string
  HeldSlotName(string) string
  SetHeld(string, thing.Thing)
  IsHolding(thing.Thing) bool
}

type Bodied interface {
  Body() Interface
}

// BasicBodyParts maps the strings identifying slots to the descriptions of
// "how" things are worn in/on that slot. The "{pp}" is intended to be
// formatted to the appropriate possessive pronoun (his, your, its, &c.) by
// the gstrings package used by other dta5 packages.
//
var BasicBodyParts = map[string]string {
  
  // humanoid body parts
  
  "head": "on {pp} head",
  "face": "on {pp} face",
  "neck": "around {pp} neck",
  "shoulder": "over {pp} shoulder",
  "drape": "across {pp} shoulders",
  "backpack": "on {pp} back",
  "shirt": "on {pp} torso",
  // nothing for "armor"
  "belt": "around {pp} waist",
  "hands": "on {pp} hands",
  "pants": "on {pp} legs",
  "feet": "on {pp} feet",
  // nothing for "misc"
  
  // nonhumanoid body parts may follow
  
}

var BasicHeldParts = map[string]string {
  // human hands
  "left_hand": "in {pp} left hand",
  "right_hand": "in {pp} right hand",
  
  // nonhuman appendages (like tentacles?) may follow
}

// BasicBody as of yet is the only type that implements the body.Interface,
// and may only ever be. It holds two maps: one mappng tokens representing
// locations a thing.Wearable could be worn to how many items can occupy each
// slot (most commonly a single item), and another mapping tokens
// representing body parts that can hold things to the things held there.
//
// When combat enters the picture, the BasicBody will probably hold a plethora
// of information about the abilities of a given Bodied individual.
//
type BasicBody struct {
  wearSlots map[string] byte
  holdSlots map[string] thing.Thing
}

// WornSlots() returns the number of items that can be worn on the body part
// represented by the supplied slot string; in most cases, this will be 0 or 1.
// The returned boolean indicates whether the BasicBody has that body part.
//
func (bb BasicBody) WornSlots(slot string) (byte, bool) {
  n, has_slot := bb.wearSlots[slot]
  return n, has_slot
}
// Return a slice of all the slot strings where things can be worn on this body.
//
func (bb BasicBody) WornSlotKeys() []string {
  slots := make([]string, 0, len(bb.wearSlots))
  for k, _ := range bb.wearSlots {
    slots = append(slots, k)
  }
  return slots
}

// Return the descriptive string corresponding to how things are worn on the
// supplied slot. Ex:
//  bb.WornSlotName("head") => "on {pp} head".
func (bb BasicBody) WornSlotName(slot string) string {
  return BasicBodyParts[slot]
}

// Return what thing.Thing is held in the holding appendage represented by the
// supplied slot string, or nil if it's empty. The returned boolean indicates
// whether the BasicBody has that body part.
//
func (bb BasicBody) HeldIn(slot string) (thing.Thing, bool) {
  t, has_slot := bb.holdSlots[slot]
  return t, has_slot
}

// Return a slice of all the slot strings where things can be held by this body.
//
func (bb BasicBody) HeldSlotKeys() []string {
  slots := make([]string, 0, len(bb.holdSlots))
  for k, _ := range bb.holdSlots {
    slots = append(slots, k)
  }
  return slots
}

// Return the descriptive string corresponding to how things are held by the
// supplied slot. Ex:
//  bb.HeldSlotName("right_hand") => "in {pp} right hand"
//
func (bb BasicBody) HeldSlotName(slot string) string {
  return BasicHeldParts[slot]
}

// Marks the supplied slot as holding the supplied thing.Thing. Caveat: you
// can unwittingly give people tentacles this way if you're not careful.
//
func (bbp *BasicBody) SetHeld(slot string, t thing.Thing) {
  bbp.holdSlots[slot] = t
}

// Return true if the supplied thing.Thing is currently being held by the body.
//
func (bb BasicBody) IsHolding(t thing.Thing) bool {
  for _, v := range bb.holdSlots {
    if v == t {
      return true
    }
  }
  return false
}

// Create and return a *BasicBody with all the appropriate slots for a humanoid
// creature.
func NewHumaniod() *BasicBody {
  nb := BasicBody{
    wearSlots: map[string]byte{
      "head": 1,
      "face": 1,
      "neck": 3,
      "shoulder": 2,
      "drape": 1,
      "backpack": 1,
      "shirt": 1,
      "armor": 1,
      "belt": 1,
      "hands": 1,
      "pants": 1,
      "feet": 1,
      "misc": 16,
    },
    holdSlots: map[string]thing.Thing {
      "right_hand": nil,
      "left_hand": nil,
    },
  }
  return &nb
}
    
