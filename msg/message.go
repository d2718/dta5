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

type Env struct {
  Type string
  Text string
}

type Message struct {
  Dir map[Messageable]Env
  Gen Env
}

func New(typ, fmtstr string, args ...interface{}) *Message {
  gen := Env{ Type: typ, Text: fmt.Sprintf(fmtstr, args...), }
  return &Message{  Dir: make(map[Messageable]Env),
                    Gen: gen, }
}

func (m *Message) Add(targ Messageable, typ, fmtstr string, args ...interface{}) {
  m.Dir[targ] = Env{ Type: typ, Text: fmt.Sprintf(fmtstr, args...), }
}
