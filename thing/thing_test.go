// thing_test.go
//
// Test suite for dta5/thing
//
package thing

import( "fmt"; "testing";
)

func TestThingList(t *testing.T) {
  tl := NewThingList(VT_UNLTD, VT_UNLTD, nil, 0)
  i0 := NewItem("i0", "a first thingy", "", false, nil, 1.0, 1.0)
  i1 := NewItem("i1", "a/an second item", "with a longer description", false, nil, 1.0, 2.0)
  i2 := NewItem("i2", "a lump", "", false, nil, 2.0, 0.5)
  
  fmt.Printf("%s\n", tl.EnglishList())
  tl.Add(i0)
  fmt.Printf("%s\n", tl.EnglishList())
  tl.Add(i1)
  fmt.Printf("%s\n", tl.EnglishList())
  tl.Add(i2)
  fmt.Printf("%s\n", tl.EnglishList())
  tl.Remove(i0)
  fmt.Printf("%s\n", tl.EnglishList())
}
