// message.go
//
// Some abstractions to facilitate "messaging", that is, text to be displayed
// describing things that are witnessed in the game world.
//
// updated 2017-08-30
//
// As of this writing, the only things for which textual messaging matters
// are players (pc.PlayerChar is the only msg.Messageable implementor whose
// Deliver() method does anything), but I hope eventually to include
// automated game constructs that can react to messaging.
//
package msg

import "fmt"

type Messageable interface {
  Deliver(*Message)
}
// The Env is a sort of "typed" message. All game messaging that should be
// delivered is "typed", so deliverees can react to it differently (for
// instance, the client program can apply different coloring or formatting).
//
type Env struct {
  Type string
  Text string
}

// A Message is meant to be delivered to an entire room.Room, describing some
// action or event taking place there. Because different witnesses will
// perceive events differently (a person being shot, for example, will see
// something different than a third party witnessing the shooting), a single
// Message may contain several Envs intended for different receivers.
//
// Message.Dir is a map associating (pointers to) specific observers with
// the Envs they should receive. Message.Gen is an Env intended for anyone
// who should receive the Message but doesn't need specialized messaging.
//
// Using the shooting example from above, a Message describing px shooting py
// would probably have the following structure:
//
//  Message{
//    Dir: { *px: Env{Type: "txt", Text: "You shoot Person Y!"},
//           *py: Env{Type: "txt", Text: "Person X shoots you!"},
//         },
//    Gen: Env{Type: "txt", Text: "Person X shoots Person Y!"},
//  }
//
type Message struct {
  Dir map[Messageable]Env
  Gen Env
}

// New() creates a new Message with the supplied Gen type and text. Calling
// style is similar to the fmt.Xprintf() functions.
//
func New(typ, fmtstr string, args ...interface{}) *Message {
  gen := Env{ Type: typ, Text: fmt.Sprintf(fmtstr, args...), }
  return &Message{  Dir: make(map[Messageable]Env),
                    Gen: gen, }
}

// Add() adds a given specific targetted Env to a message. Calling style is
// similar to the fmt.Xprintf() functions.
//
func (m *Message) Add(targ Messageable, typ, fmtstr string, args ...interface{}) {
  m.Dir[targ] = Env{ Type: typ, Text: fmt.Sprintf(fmtstr, args...), }
}
