// pc.go
//
// dta5 Player Character stuff
//
// updated 2017-08-08
//
package pc

import( "encoding/json"; "fmt"; "net"; "os"; "path/filepath"; "sync"; "time";
        "golang.org/x/crypto/bcrypt";
        "dta5/act"; "dta5/body"; "dta5/desc";
        "dta5/name"; "dta5/load"; "dta5/log"; "dta5/msg"; "dta5/ref";
        "dta5/room"; "dta5/save"; "dta5/thing";
)

var PlayerDir string = "pc_dir"
var HashCost int = 4
var ClientVersion int = 170807

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("pc.go: " + fmtstr, args...))
}

type PlayerState struct {
  RefToken  string
  PassHash  string
  NameTitle string
  NameFirst string
  NameRest  string
  RightHand string
  LeftHand  string
  name.Gender
  Location  string
  Inventory []string
}

type PM struct {
  Type string
  Payload string
}

const INV byte = 0

type PlayerChar struct {
  ref       string
  uname     string
  passHash  string
  name.ProperName
  where     thing.LocVec
  Inventory *thing.ThingList
  bod       *body.BasicBody
  conn      net.Conn
  rcvr      *json.Decoder
  sndr      *json.Encoder
  sndlockr  *sync.Mutex
}

func (p PlayerChar) Ref() string { return p.ref }
func (p PlayerChar) Mass() thing.TVal { return thing.INFTY }
func (p PlayerChar) Bulk() thing.TVal { return thing.INFTY }
func (p PlayerChar) Loc() thing.LocVec { return p.where }
func (p PlayerChar) Body() body.Interface { return p.bod }

func (pp *PlayerChar) SetLoc(loc thing.LocVec) {
  pp.where = loc
}

func (p PlayerChar) Data(tok string) interface{} {
  if thing.Data[p.ref] == nil {
    return nil
  } else {
    return thing.Data[p.ref][tok]
  }
}

func (p PlayerChar) SetData(tok string, val interface{}) {
  if thing.Data[p.ref] == nil {
    thing.Data[p.ref] = make(map[string]interface{})
  }
  thing.Data[p.ref][tok] = val
}

var PlayerChars = make(map[string]*PlayerChar)

func Login(newConn net.Conn) error {
  
  log(dtalog.DBG, "Login() called")
  
  new_rcvr := json.NewDecoder(newConn)
  new_sndr := json.NewEncoder(newConn)
  
  var mesg PM = PM{ Type: "version", Payload: "experimental", }
  err := new_sndr.Encode(mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error sending version message: %s", err)
    mesg = PM{ Type: "logout", Payload: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): version message sent")
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding version message: %s", err)
    mesg = PM{ Type: "logout", Payload: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  
  if mesg.Type != "version" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := PM{ Type: "logout", Payload: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  
  var version int
  fmt.Sscanf(mesg.Payload, "%d", &version)
  if version < ClientVersion {
    log(dtalog.MSG, "Login(): client (version %d) out of date, rejecting login", version)
    reply := PM{ Type: "logout", Payload: "updated client required", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("client update required")
  }
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding uname message: %s", err)
    mesg = PM{ Type: "logout", Payload: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): response rec'd: %v", mesg)
  
  if mesg.Type != "uname" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := PM{ Type: "logout", Payload: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  uname := mesg.Payload
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding pwd message: %s", err)
    mesg = PM{ Type: "logout", Payload: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): response rec'd: %v", mesg)
  
  if mesg.Type != "pwd" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := PM{ Type: "logout", Payload: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  pwd := mesg.Payload
  
  plr_path := filepath.Join(PlayerDir, uname + ".json")
  f, err := os.Open(plr_path)
  if err != nil {
    log(dtalog.ERR, "Login(): error opening file %q: %s", plr_path, err)
    reply := PM{ Type: "logout", Payload: fmt.Sprintf("unable to login %q", uname), }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("cannot open file %q", plr_path)
  }
  log(dtalog.DBG, "Login(): opened player file")
  
  var ps PlayerState
  psdcdr := json.NewDecoder(f)
  err = psdcdr.Decode(&ps)
  
  if err != nil {
    log(dtalog.ERR, "Login(): error reading file %q: %s", plr_path, err)
    reply := PM{ Type: "logout", Payload: "there was an error", }
    new_sndr.Encode(reply)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): read saved player state %v", ps)
  
  if bcrypt.CompareHashAndPassword([]byte(ps.PassHash), []byte(pwd)) != nil {
    log(dtalog.MSG, "Login(): hashed password does not match")
    reply := PM{ Type: "logout", Payload: "username and password don't match", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("username and password don't match")
  }
  
  if ref.Deref(ps.RefToken) != nil {
    log(dtalog.MSG, "Login(): player %q already logged in", uname)
    reply := PM{ Type: "logout", Payload: fmt.Sprintf("user %q already logged in", uname), }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("user already logged in")
  }
  
  new_pc := PlayerChar{
    ref: ps.RefToken,
    uname: uname,
    ProperName: name.ProperName{ Title: ps.NameTitle, First: ps.NameFirst,
                                 Rest: ps.NameRest, Gender: ps.Gender },
    Inventory: thing.NewThingList(thing.VT_UNLTD, thing.VT_UNLTD, nil, 0),
    passHash: ps.PassHash,
    bod:  body.NewHumaniod(),
    conn: newConn,
    rcvr: new_rcvr,
    sndr: new_sndr,
    sndlockr: new(sync.Mutex),
  }
  log(dtalog.DBG, "Login(): created PlayerChar struct")
  
  ref.Register(&new_pc)
  new_pc.Inventory.LocVec = thing.LocVec{ Place: &new_pc, Side: INV, }
  for psdcdr.More() {
    var x []interface{}
    err = psdcdr.Decode(&x)
    if err != nil {
      log(dtalog.ERR, "Login(): error decoding inventory: %s", err)
    } else {
      load.LoadFeature(x, load.MUT)
    }
  }
  for _, t_id := range ps.Inventory {
    new_pc.Inventory.Add(ref.Deref(t_id).(thing.Thing))
  }
  unlimbo_func := func (t thing.Thing) {
    desc.UnLimbo(t)
  }
  new_pc.Inventory.Walk(unlimbo_func)
  
  if ps.RightHand != "" {
    new_pc.Body().SetHeld("right_hand", ref.Deref(ps.RightHand).(thing.Thing))
  }
  if ps.LeftHand != "" {
    new_pc.Body().SetHeld("left_hand", ref.Deref(ps.LeftHand).(thing.Thing))
  }
  
  log(dtalog.DBG, "Login(): registered and loaded inventory")
  
  start_room := ref.Deref(ps.Location).(*room.Room)
  start_room.Contents.Add(&new_pc)
  log(dtalog.DBG, "Login(): added to Room")
  
  arrive_msg := msg.New(fmt.Sprintf("%s arrives.", new_pc.Normal(0)))
  arrive_msg.Add(&new_pc, "You arrive.")
  
  greet_act := act.Action{
    Time: time.Now(),
    Act: func() error {
      start_room.Deliver(arrive_msg)
      DoLook(&new_pc, "look", nil, "", nil, "")
      return nil
    },
  }
  
  act.Enqueue(&greet_act)
  log(dtalog.DBG, "Login(): Enqueued arrival message")
  
  PlayerChars[new_pc.ref] = &new_pc
  go (&new_pc).listen()
  
  log(dtalog.DBG, "Login() returning")
  return nil
}

func (pp *PlayerChar) Logout() error {
  state := PlayerState{
    PassHash:  pp.passHash,
    RefToken:  pp.Ref(),
    NameTitle: pp.ProperName.Title,
    NameFirst: pp.ProperName.First,
    NameRest:  pp.ProperName.Rest,
    Gender:    pp.ProperName.Gender,
    Location:  pp.where.Place.Ref(),
    Inventory: make([]string, 0, len(pp.Inventory.Things)),
  }
  
  bod := pp.Body()
  
  if rh, _ := bod.HeldIn("right_hand"); rh != nil {
    state.RightHand = rh.Ref()
  }
  if lh, _ := bod.HeldIn("left_hand"); lh != nil {
    state.LeftHand = lh.Ref()
  }
  
  s, err := save.New(filepath.Join(PlayerDir, pp.uname + ".json"))
  if err != nil {
    log(dtalog.ERR, "(*PlayerChar %q) Logout(): unable to open save file: %s",
                    pp.Short(0), err)
    return err
  }
  defer s.Close()
  
  for _, t := range pp.Inventory.Things {
    state.Inventory = append(state.Inventory, t.Ref())
  }
  s.Encode(state)
  
  for _, t := range pp.Inventory.Things {
    t.Save(*s)
    ref.Deregister(t)
  }
  
  loc := pp.where.Place.(*room.Room)
  m := msg.New(fmt.Sprintf("%s leaves.", pp.Normal(0)))
  m.Add(pp, "You leave.")
  loc.Deliver(m)
  loc.Contents.Remove(pp)
  pp.Send(PM{ Type: "logout", Payload: "You have been logged out.", })
  pp.conn.Close()
  delete(PlayerChars, pp.ref)
  ref.Deregister(pp)
  
  return nil
}

func (pp *PlayerChar) Save(save.Saver) {
  return
}
  
// listens to connection; deals with incoming commands
//
func (pp *PlayerChar) listen() {
  for {
    var cmd PM
    err := pp.rcvr.Decode(&cmd)
    if err != nil {
      log(dtalog.ERR, "(*PlayerChar %q) listen(): error receiving: %s",
          pp.Short(0), err)
      return
    } else if cmd.Type == "cmd" {
      a := act.Action{
        Time: time.Now(),
        Act: func() error { return pp.Parse(cmd.Payload)},
      }
      act.Enqueue(&a)
    }
  }
}

func (p PlayerChar) Send(pm PM) {
  p.sndlockr.Lock()
  p.sndr.Encode(pm)
  p.sndlockr.Unlock()
}

func (pp* PlayerChar) QWrite(fmtstr string, args ...interface{}) {
  txt := fmt.Sprintf(fmtstr, args...)
  pp.Send(PM{ Type: "txt", Payload: txt, })
} 

func (pp *PlayerChar) Deliver(m *msg.Message) {
  txt, ok := m.Dir[pp]
  if !ok {
    txt = m.Gen
  }
  if len(txt) > 0 {
    pp.Send(PM{ Type: "txt", Payload: txt})
  }
}

func (pp *PlayerChar) React(cmd string) {
  log(dtalog.DBG, "(*PlayerChar %q) React(): reacting: %q", pp.Short(0), cmd)
  
  if cmd == "quit" {
    pp.Logout()
    return
  }
  
  m := msg.New(fmt.Sprintf("%s does this: %s", pp.Normal(0), cmd))
  m.Add(pp, fmt.Sprintf("You do this: %s", cmd))
  a := act.Action{
    Time: time.Now(),
    Act: func() error {
      pp.where.Place.(*room.Room).Deliver(m)
      return nil
    },
  }
  act.Enqueue(&a)
}

// AllButMe() returns a thing.ThingList containing the PlayerChar's current
// room.Room's.Contents but without the PlayerChar.
//
// This is useful for finding the objects of verbs.
//
func (pp *PlayerChar) AllButMe() *thing.ThingList {
  src := pp.where.Place.(*room.Room).Contents
  nts := make([]thing.Thing, 0, len(src.Things))
  
  for _, t := range src.Things {
    if t != pp {
      nts = append(nts, t)
    }
  }
  
  ntl := thing.ThingList{ Things: nts, }
  return &ntl
}

func (p PlayerChar) InHand(obj thing.Thing) bool {
  b := p.Body()
  if t, _ := b.HeldIn("right_hand"); t == obj {
    return true
  } else if t, _ := b.HeldIn("left_hand"); t == obj {
    return true
  } else {
    return false
  }
}

func (pp PlayerChar) SetDescPage(sp *string) { return }
func (pp PlayerChar) Desc() string { return fmt.Sprintf("You see %s.", pp.Full(0)) }

func Wall(pm PM) {
  for _, pp := range PlayerChars {
    pp.Send(pm)
  }
}
