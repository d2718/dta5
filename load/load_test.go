//
package load

import( "fmt"; "os"; "testing";
        "dta5/log"; "dta5/ref"; //"dta5/thing";
)

func init() {
  dtalog.Start(dtalog.ERR, os.Stderr)
  dtalog.Start(dtalog.DBG, os.Stderr)
  dtalog.Start(dtalog.MSG, os.Stderr)
}

func TestLoading(t *testing.T) {
  err := LoadFile("tworld.json")
  if err != nil {
    t.Errorf("error loading file\n")
  }
  
  for k, v := range ref.Referents {
    fmt.Printf("%q: %v\n", k, v)
  }
}
