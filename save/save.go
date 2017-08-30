// save.go
//
// An almost-unnecessary abstraction for saving the state of the game.
//
// updated 2017-08-30
//
// A Saver is really no more than a pointer to the save game os.File and
// a json.Encoder that writes to that file. Any saveable object (which should
// implement the save.Interface) should know how to generate the appropriate
// JSON object to represent itself, and its Save() method should ask the
// supplied Saver to encode that object.
//
package save

import( "encoding/json"; "os"; )

type Saver struct {
  *os.File
  *json.Encoder
}

type Interface interface {
  Save(Saver)
  Ref() string
}

func New(pth string) (*Saver, error) {
  f, err := os.OpenFile(pth, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
  if err != nil {
    return nil, err
  }
  
  ncdr := json.NewEncoder(f)
  
  return &Saver{ File: f, Encoder: ncdr, }, nil
}
