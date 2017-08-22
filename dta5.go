// dta5.go
//
// The FIFTH try at a text adventure. The... thee-and-a-halfth? Try at
// an online multi-player environment.
//
// This file contains the main function.
//
// updated 2017-08-08
//
package main

import( "bufio"; "fmt"; "flag"; "net"; "os"; "path/filepath"; "strings";
        "time";
        "github.com/d2718/dconfig";
        "dta5/log";
        "dta5/act"; "dta5/desc"; "dta5/door"; "dta5/load"; "dta5/mood";
        "dta5/msg"; "dta5/pc"; "dta5/ref"; "dta5/room"; "dta5/scripts";
        "dta5/scripts/more"; "dta5/save";
)

const mainWorldFile = "main.json"
const descPath      = "descs"

var sockName string = "ctrl"
var worldDir string
var actionQueueLength int = 256
var listenPort string = ":10102"
var unloadInterval = time.Duration(15) * time.Second

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("dta5.go: " + fmtstr, args...))
}

func Configure(cfgPath string) {
  var port_cfgint  int = 10102
  var stale_cfgint int = 300 
  
  dconfig.Reset()
  dconfig.AddInt(&actionQueueLength, "queue_length", dconfig.UNSIGNED)
  dconfig.AddInt(&port_cfgint,       "port",         dconfig.UNSIGNED)
  dconfig.AddInt(&stale_cfgint,      "page_life",    dconfig.UNSIGNED)
  dconfig.AddString(&pc.HelpDir,     "help_dir",     dconfig.STRIP)
  dconfig.AddInt(&pc.HashCost,       "hash_cost",    dconfig.UNSIGNED)
  dconfig.Configure([]string{cfgPath}, true)
  
  listenPort = fmt.Sprintf(":%d", port_cfgint)
  unloadInterval = time.Duration(stale_cfgint) * time.Second
  desc.StalePageLife = time.Duration(stale_cfgint) * time.Second
}

func listenForConnections() {
  lsnr, err := net.Listen("tcp", listenPort)
  if err != nil {
    log(dtalog.ERR, "listenForConnections(): error in net.Listen(): %s\n", err)
    os.Exit(1)
  }
  
  for {
    new_conn, err := lsnr.Accept()
    if err != nil {
      log(dtalog.ERR, "listenForConnection(): error in (net.Listener()) Accept(): %s\n", err)
      os.Exit(1)
    }
    err = pc.Login(new_conn)
    if err != nil {
      log(dtalog.ERR, "ListenForConnection(): error in pc.Login(): %s\n", err)
    }
  }
}

var commandChannel = make(chan string)

func listenToStdin() {
  scnr := bufio.NewScanner(os.Stdin)
  for scnr.Scan() {
    commandChannel <- scnr.Text()
  }
}

func listenOnSocket() {
  lsnr, err := net.ListenUnix("unix", &net.UnixAddr{sockName, "unix"})
  if err != nil {
    log(dtalog.ERR, "listenOnSocket(): error opening listener: %s\n", err)
    os.Exit(1)
  }
  
  for run {
    conn, err := lsnr.AcceptUnix()
    if err != nil {
      log(dtalog.ERR, "listenOnSocket(): error accepting connection: %s\n", err)
    } else {
      scnr := bufio.NewScanner(conn)
      for scnr.Scan() {
        commandChannel <- scnr.Text()
      }
    }
    conn.Close()
  }
  
  lsnr.Close()
}

func autoUnload() error {
  desc.UnloadStale()
  again := act.Action{
    Time: time.Now().Add(unloadInterval),
    Act: autoUnload,
  }
  act.Enqueue(&again)
  return nil
}
    
func processCommand(cmd string) {
  log(dtalog.DBG, "processCommand(): rec'd: %q", cmd)
  cmd_slice := strings.SplitN(cmd, " ", 2)
  var verb, rest string
  
  verb = strings.ToLower(cmd_slice[0])
  if len(cmd_slice) > 1 {
    rest = cmd_slice[1]
  }
  
  switch verb {
  case "wall":
    pc.Wall(msg.Env{Type: "sys", Text: rest})
    
  case "save":
    if rest == "" {
      log(dtalog.MSG, "processCommand(): you must specify an identifier to save.")
      return
    }
    m := msg.Env{
      Type: "sys",
      Text: "You are being logged out so that the state of the game may be saved.",
    }
    pc.Wall(m)
    for _, pp := range pc.PlayerChars {
      pp.Logout()
    }
    save_path := filepath.Join(worldDir, "saves", rest + ".json")
    s, err := save.New(save_path)
    if err != nil {
      log(dtalog.ERR, "processCommand(): error creating save.Saver: %s", err)
      return
    }
    defer s.Close()
    
    save_func := func(r ref.Interface) {
      switch t_r := r.(type) {
      case *room.Room:
        t_r.Save(*s)
      default:
        return
      }
    }

    ref.Walk(save_func)
    for _, d := range door.Doors {
      d.Save(*s)
    }
    ref.SaveData(*s)
    scripts.SaveBindings(*s)
    
    log(dtalog.DBG, "processCommand(): save complete")
  
  case "load":
    if rest == "" {
      log(dtalog.MSG, "processCommand(): you must specify and identifier to load.")
      return
    }
    m := msg.Env{
      Type: "sys",
      Text: "You are being logged out so that the state of the game may be loaded.",
    }
    pc.Wall(m)
    for _, pp := range pc.PlayerChars {
      pp.Logout()
    }
    load_path := filepath.Join(worldDir, "saves", rest + ".json")
    ref.Reset()
    door.Reset()
    mood.Initialize()
    load.LoadFile(filepath.Join(worldDir, mainWorldFile), load.PERM)
    load.LoadFile(load_path, load.MUT)
    desc.Initialize(filepath.Join(worldDir, descPath))
    act.Initialize(actionQueueLength)
    first_unload := act.Action{
      Time: time.Now().Add(unloadInterval),
      Act: autoUnload,
    }
    act.Enqueue(&first_unload)
    for _, mp := range mood.Messengers {
      mp.Arm()
    }
    
    log(dtalog.DBG, "processCommand(): load complete")
    
  case "quit":
    m := msg.Env{
      Type:"sys",
      Text: "The game is being shut down RIGHT NOW.",
    }
    pc.Wall(m)
    for _, pp := range pc.PlayerChars {
      pp.Logout()
    }
    run = false
  default:
    log(dtalog.MSG, "processCommand(): unrecognized command: %q", verb)
  }
}

var run bool = true

func main() {
  flag.Parse()
  worldDir = flag.Arg(0)
  load.WorldDir = worldDir
  
  dtalog.Start(dtalog.ERR, os.Stdout)
  dtalog.Start(dtalog.WRN, os.Stdout)
  dtalog.Start(dtalog.MSG, os.Stdout)
  //dtalog.Start(dtalog.DBG, os.Stdout)
  
  Configure(filepath.Join(worldDir, "conf"))
  pc.PlayerDir = filepath.Join(worldDir, "pc_dir")
  
  more.Initialize()
  mood.Initialize()
  load.LoadFile(filepath.Join(worldDir, mainWorldFile), load.INIT)
  desc.Initialize(filepath.Join(worldDir, descPath))
  act.Initialize(actionQueueLength)
  first_unload := act.Action{
    Time: time.Now().Add(unloadInterval),
    Act: autoUnload,
  }
  act.Enqueue(&first_unload)
  for _, mp := range mood.Messengers {
    mp.Arm()
  }
  
  go listenForConnections()
  // go listenToStdin()
  go listenOnSocket()
  
  for run {
    select {
    case cmd := <- commandChannel:
      processCommand(cmd)
    default:
      a := act.Next()
      if a == nil {
        time.Sleep(100 * time.Millisecond)
      } else {
        err := a.Act()
        if err != nil {
          log(dtalog.ERR, "main(): run error: %s\n", err)
        }
      }
    }
  }
}

