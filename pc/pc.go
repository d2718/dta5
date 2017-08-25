// pc.go
//
// dta5 Player Character stuff
//
// updated 2017-08-18
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
var ClientVersion int = 170818

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("pc: " + fmtstr, args...))
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

func (p PlayerChar) Data(key string) interface{} {
  return ref.GetData(p, key)
}
func (p PlayerChar) SetData(key string, val interface{}) {
  ref.SetData(p, key, val)
}

var PlayerChars = make(map[string]*PlayerChar)

func Login(newConn net.Conn) error {
  
  log(dtalog.DBG, "Login() called")
  
  new_rcvr := json.NewDecoder(newConn)
  new_sndr := json.NewEncoder(newConn)
  
  var mesg msg.Env = msg.Env{ Type: "version",
                              Text: fmt.Sprintf("%d", ClientVersion), }
  err := new_sndr.Encode(mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error sending version message: %s", err)
    mesg = msg.Env{ Type: "logout", Text: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): version message sent")
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding version message: %s", err)
    mesg = msg.Env{ Type: "logout", Text: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  
  if mesg.Type != "version" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := msg.Env{ Type: "logout", Text: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  
  var version int
  fmt.Sscanf(mesg.Text, "%d", &version)
  if version < ClientVersion {
    log(dtalog.MSG, "Login(): client (version %d) out of date, rejecting login", version)
    reply := msg.Env{ Type: "logout", Text: "updated client required", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("client update required")
  }
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding uname message: %s", err)
    mesg = msg.Env{ Type: "logout", Text: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): response rec'd: %v", mesg)
  
  if mesg.Type != "uname" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := msg.Env{ Type: "logout", Text: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  uname := mesg.Text
  
  err = new_rcvr.Decode(&mesg)
  if err != nil {
    log(dtalog.ERR, "Login(): error decoding pwd message: %s", err)
    mesg = msg.Env{ Type: "logout", Text: "communication error", }
    new_sndr.Encode(mesg)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): response rec'd: %v", mesg)
  
  if mesg.Type != "pwd" {
    log(dtalog.MSG, "Login(): incorrect protocol from client")
    reply := msg.Env{ Type: "logout", Text: "incorrect login protocol", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("incorrect login protocol: %v", mesg)
  }
  pwd := mesg.Text
  
  plr_path := filepath.Join(PlayerDir, uname + ".json")
  f, err := os.Open(plr_path)
  if err != nil {
    log(dtalog.ERR, "Login(): error opening file %q: %s", plr_path, err)
    reply := msg.Env{ Type: "logout", Text: fmt.Sprintf("unable to login %q", uname), }
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
    reply := msg.Env{ Type: "logout", Text: "there was an error", }
    new_sndr.Encode(reply)
    newConn.Close()
    return err
  }
  log(dtalog.DBG, "Login(): read saved player state %v", ps)
  
  if bcrypt.CompareHashAndPassword([]byte(ps.PassHash), []byte(pwd)) != nil {
    log(dtalog.MSG, "Login(): hashed password does not match")
    reply := msg.Env{ Type: "logout", Text: "username and password don't match", }
    new_sndr.Encode(reply)
    newConn.Close()
    return fmt.Errorf("username and password don't match")
  }
  
  if ref.Deref(ps.RefToken) != nil {
    log(dtalog.MSG, "Login(): player %q already logged in", uname)
    reply := msg.Env{ Type: "logout", Text: fmt.Sprintf("user %q already logged in", uname), }
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
  
  arrive_msg := msg.New("txt", fmt.Sprintf("%s arrives.", new_pc.Normal(0)))
  arrive_msg.Add(&new_pc, "txt", "You arrive.")
  
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

func (pp *PlayerChar) Logout(mesg string) error {
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
  m := msg.New("txt", fmt.Sprintf("%s leaves.", pp.Normal(0)))
  m.Add(pp, "txt", "You leave.")
  loc.Deliver(m)
  loc.Contents.Remove(pp)
  pp.Send(msg.Env{ Type: "logout", Text: mesg, })
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
    var cmd msg.Env
    err := pp.rcvr.Decode(&cmd)
    if err != nil {
      log(dtalog.ERR, "(*PlayerChar %q) listen(): error receiving: %s",
          pp.Short(0), err)
      return
    } else if cmd.Type == "cmd" {
      act.Add(0.0, func() error { return pp.Parse(cmd.Text) })
    }
  }
}

func (p PlayerChar) Send(pm msg.Env) {
  p.sndlockr.Lock()
  p.sndr.Encode(pm)
  p.sndlockr.Unlock()
}

func (pp* PlayerChar) QWrite(fmtstr string, args ...interface{}) {
  txt := fmt.Sprintf(fmtstr, args...)
  pp.Send(msg.Env{ Type: "txt", Text: txt, })
} 

func (pp *PlayerChar) Deliver(m *msg.Message) {
  nvlp, ok := m.Dir[pp]
  if !ok {
    nvlp = m.Gen
  }
  if len(nvlp.Text) > 0 {
    pp.Send(nvlp)
  }
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

func Wall(pm msg.Env) {
  for _, pp := range PlayerChars {
    pp.Send(pm)
  }
}
