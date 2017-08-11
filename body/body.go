// body.go
//
// dta5 Body struct and methods
// Bodies are for Things (like PlayerChars) that can wear clothes and
// participate in combat.
//
// updated 2017-08-11
//
package body

import( "fmt";
        "dta5/log"; "dta5/thing"
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("body: " + fmtstr, args...))
}

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
  "left_hand": "in {pp} left hand",
  "right_hand": "in {pp} right hand",
}

type BasicBody struct {
  wearSlots map[string] byte
  holdSlots map[string] thing.Thing
}

func (bb BasicBody) WornSlots(slot string) (byte, bool) {
  n, has_slot := bb.wearSlots[slot]
  return n, has_slot
}
func (bb BasicBody) WornSlotKeys() []string {
  slots := make([]string, 0, len(bb.wearSlots))
  for k, _ := range bb.wearSlots {
    slots = append(slots, k)
  }
  return slots
}

func (bb BasicBody) WornSlotName(slot string) string {
  return BasicBodyParts[slot]
}

func (bb BasicBody) HeldIn(slot string) (thing.Thing, bool) {
  t, has_slot := bb.holdSlots[slot]
  return t, has_slot
}
func (bb BasicBody) HeldSlotKeys() []string {
  slots := make([]string, 0, len(bb.holdSlots))
  for k, _ := range bb.holdSlots {
    slots = append(slots, k)
  }
  return slots
}
func (bb BasicBody) HeldSlotName(slot string) string {
  return BasicHeldParts[slot]
}

func (bbp *BasicBody) SetHeld(slot string, t thing.Thing) {
  bbp.holdSlots[slot] = t
}

func (bb BasicBody) IsHolding(t thing.Thing) bool {
  for _, v := range bb.holdSlots {
    if v == t {
      return true
    }
  }
  return false
}

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
    
