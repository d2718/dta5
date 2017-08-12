// room.go
//
// dta5 Room struct and methods
//
// updated 2017-08-02
//
package room

import(
        "dta5/desc"; "dta5/msg"; "dta5/thing"; "dta5/ref"; "dta5/save";
)

const(  SCENERY byte = 0
        CONTENTS byte = 1
)

type NavDir int
const(  N     NavDir = NavDir(0)
        NE    NavDir = NavDir(1)
        E     NavDir = NavDir(2)
        SE    NavDir = NavDir(3)
        S     NavDir = NavDir(4)
        SW    NavDir = NavDir(5)
        W     NavDir = NavDir(6)
        NW    NavDir = NavDir(7)
        UP    NavDir = NavDir(8)
        DOWN  NavDir = NavDir(9)
        OUT   NavDir = NavDir(10)
)

var NavDirNames map[NavDir]string = map[NavDir]string {
  N: "north", NE: "northeast", E: "east", SE: "southeast", S: "south",
  SW: "southwest", W: "west", NW: "northwest", UP: "up", DOWN: "down",
  OUT: "out",
}

type Room struct {
  ref      string
  Title    string
  DescText string
  descPage *string
  Scenery  *thing.ThingList
  Contents *thing.ThingList
  nav      []string
}

func NewRoom(r, title string, navz ...string) *Room {
  nr := Room{ ref: r, Title: title, nav: make([]string, 11, 11) }
  err := ref.Register(&nr)
  if err != nil {
    return nil
  }
  
  nr.Scenery  = thing.NewThingList(thing.VT_UNLTD, thing.VT_UNLTD, &nr, SCENERY)
  nr.Contents = thing.NewThingList(thing.VT_UNLTD, thing.VT_UNLTD, &nr, CONTENTS)
  
  for n, x := range navz {
    nr.nav[n] = x
  }
  
  return &nr
}

func (r Room) Ref() string {
  return r.ref
}

func (r *Room) SetDescPage(npage *string) {
  r.descPage = npage
}

func (r Room) Desc() string {
  if r.descPage == nil {
    return "This space unintentionally left blank."
  }
  
  return desc.GetDesc(*(r.descPage), r.ref)
}

func (r Room) Data(key string) interface{} {
  return ref.GetData(r, key)
}
func (r Room) SetData(key string, val interface{}) {
  ref.SetData(r, key, val)
}

func (r Room) Nav(d NavDir) ref.Interface {
  x := r.nav[int(d)]
  if x == "" {
    return nil
  } else {
    return ref.Deref(x)
  }
}

func (r Room) Deliver(m *msg.Message) {
  r.Contents.Deliver(m)
}

func (r Room) ExitDirs() []NavDir {
  x := make([]NavDir, 0, 11)
  for n, d := range r.nav {
    if d != "" {
      x = append(x, NavDir(n))
    }
  }
  return x
}

func (r *Room) Save(s save.Saver) {
  var s_pop, c_pop []interface{}
  if len(r.Scenery.Things) > 0 {
    s_pop = make([]interface{}, 0, len(r.Scenery.Things) + 3)
    s_pop = append(s_pop, "pop")
    s_pop = append(s_pop, r.Ref())
    s_pop = append(s_pop, "s")
  }
  if len(r.Contents.Things) > 0 {
    c_pop = make([]interface{}, 0, len(r.Contents.Things) + 3)
    c_pop = append(c_pop, "pop")
    c_pop = append(c_pop, r.Ref())
    c_pop = append(c_pop, "c")
  }
  
  for _, t := range r.Scenery.Things {
    t.Save(s)
    s_pop = append(s_pop, t.Ref())
    
  }
  for _, t := range r.Contents.Things {
    t.Save(s)
    c_pop = append(c_pop, t.Ref())
  }
  
  if len(r.Scenery.Things) > 0 {
    s.Encode(s_pop)
  }
  if len(r.Contents.Things) > 0 {
    s.Encode(c_pop)
  }
}
