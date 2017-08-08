// save.go
//
// dta5 state-saving
//
// updated 2017-08-01
//
package save

import( "encoding/json"; "os";
)

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
