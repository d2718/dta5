// message.go
//
// For implementing messaging.
//
// updated 2017-07-22
//
package msg

import( "fmt";
)

type Messageable interface {
  Deliver(*Message)
}

type Message struct {
  Dir map[Messageable]string
  Gen string
}

func New(fmtstr string, args ...interface{}) *Message {
  gen := fmt.Sprintf(fmtstr, args...)
  return &Message{  Dir: make(map[Messageable]string),
                    Gen: gen, }
}

func (m *Message) Add(targ Messageable, fmtstr string, args ...interface{}) {
  m.Dir[targ] = fmt.Sprintf(fmtstr, args...)
}
