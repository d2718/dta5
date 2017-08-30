// dta5.go
//
// The FIFTH try at a text adventure. The... thee-and-a-halfth? Try at
// an online multi-player environment.
//
// updated 2017-08-28
//
// At any given time, this is not guaranteed to run on any platform other
// than the one I'm using as a game server.
//
// Invocation:
//
//  ./dta5 path/to/world/dir
//
// or, better yet,
//
//  nohup ./dta5 path/to/world/dir &
//
// Where path/to/world/dir is a directory that contains the base "main.json"
// file for the game world you intend to run. (See the dta5/load package for
// specifics on file format.) Right now, if you have just cloned the git
// repository and built dta5.go (with go build), you can run the development
// test world with
//
//  nohup ./dta5 _devw/dw2 &
//
// Administrative control of the game world is through a socket called "ctrl"
// that will be created in the game world. The best way to talk to the server
// from the command line is by using netcat. For example, to shut the game
// down:
//
//  echo "quit" | nc -U ctrl
//
package main

import( "bufio"; "bytes"; "fmt"; "flag"; "net"; "os"; "path/filepath";
        "strings"; "time";
        "github.com/d2718/dconfig";
        "dta5/log";
        "dta5/act"; "dta5/desc"; "dta5/door"; "dta5/load"; "dta5/mood";
        "dta5/msg"; "dta5/pc"; "dta5/ref"; "dta5/room"; "dta5/scripts";
        "dta5/scripts/more"; "dta5/save";
)

const DEBUG = false

// Filename of the initial file the game world tries to load in the specified
// world directory.
const mainWorldFile string = "main.json"
// Name of the directory in the world directory where the dta5/desc
// description files live.
const descPath string      = "descs"
// Name of the control socket through which commands are issued to the server.
const sockName string      = "ctrl"

// Directory containing the world information to be loaded; specified on
// the command line.
var worldDir string
// Size of the buffered channels used for sending to/receiving from the
// dta5/act queue of actions. This is configurable and may require tuning
// for large game worlds or high-traffic servers.
var actionQueueLength  int = 256
// Size of the buffered channel used to send commands to the server. This
// really shouldn't need to be very large; you might even be able to get
// away with an unbuffered channel.
var commandQueueLength int = 16
// The port on which the game listens for connections from clients. This
// option is configurable.
var listenPort string = ":10102"
// The time after which unused dta5/desc Pages (q.v. the package) are
// considered "stale" and unloaded. Also the time between checks for staleness.
var unloadInterval = time.Duration(15) * time.Second
// Channel which conveys commands from the socket to the main() function.
var commandChannel chan string

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("dta5.go: " + fmtstr, args...))
}

// Read configuration files and set appropriate variable values. See the
// configuration file included in the development test world
// ("_devw/dw2/conf") for explanations of the options.
//
func Configure(cfgPath string) {
  var port_cfgint  int = 10102
  var stale_cfgint int = 300 
  
  dconfig.Reset()
  dconfig.AddInt(&actionQueueLength,  "queue_length",      dconfig.UNSIGNED)
  dconfig.AddInt(&commandQueueLength, "cmd_queue_length",  dconfig.UNSIGNED)
  dconfig.AddInt(&port_cfgint,        "port",              dconfig.UNSIGNED)
  dconfig.AddInt(&stale_cfgint,       "page_life",         dconfig.UNSIGNED)
  dconfig.AddString(&pc.HelpDir,      "help_dir",          dconfig.STRIP)
  dconfig.AddInt(&pc.HashCost,        "hash_cost",         dconfig.UNSIGNED)
  dconfig.Configure([]string{cfgPath}, true)
  
  listenPort = fmt.Sprintf(":%d", port_cfgint)
  unloadInterval = time.Duration(stale_cfgint) * time.Second
  desc.StalePageLife = time.Duration(stale_cfgint) * time.Second
}

// Meant to run as a goroutine. Listens for connecting clients and attempts
// to log them in.
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

// Reads server commands from stdin and shoves them in the channel to be
// processed by main(). As of 2017-08-29 this is no longer used, because
// communication with the server is done through a socket.
//
//~ func listenToStdin() {
  //~ scnr := bufio.NewScanner(os.Stdin)
  //~ for scnr.Scan() {
    //~ commandChannel <- scnr.Text()
  //~ }
//~ }

// Listens to the command socket and shoves commands into the channel for
// processing by main().
//
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

// Unloads stale dta5/desc pages, then sets itself to fire again after the
// configured interval.
//
func autoUnload() error {
  desc.UnloadStale()
  again := act.Action{
    Time: time.Now().Add(unloadInterval),
    Act: autoUnload,
  }
  act.Enqueue(&again)
  return nil
}

// processCommand() takes appropriate action when commands come in through
// the command socket.
//
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
  
  case "wallfile":
    f, err := os.Open(rest)
    if err != nil {
      log(dtalog.MSG, "processCommand(): unable to open file %q: %s", rest, err)
      return
    }
    defer f.Close()
    buff := new(bytes.Buffer)
    buff.ReadFrom(f)
    pc.Wall(msg.Env{Type: "sys", Text: buff.String()})
    
  case "save":
    if rest == "" {
      log(dtalog.MSG, "processCommand(): you must specify an identifier to save.")
      return
    }
    for _, pp := range pc.PlayerChars {
      pp.Logout("You have been logged out so that the state of the game may be saved.")
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
    for _, pp := range pc.PlayerChars {
      pp.Logout("You are being logged out so that the state of the game may be loaded.")
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
  
  case "logout":
    lo_slice := strings.SplitN(rest, " ", 2)
    if len(lo_slice) < 2 {
      fmt.Printf("Not enough arguments to the \"logout\" command.\n")
    } else {
      r_idx, reason := lo_slice[0], lo_slice[1]
      r := ref.Deref(r_idx)
      if pp, ok := r.(*pc.PlayerChar); ok {
        pp.Logout(reason)
      } else {
        fmt.Printf("%q is not the reference ID of a logged-in player.\n", r_idx)
      }
    }
    
  case "quit":
    for _, pp := range pc.PlayerChars {
      pp.Logout("The game has been shut down.")
    }
    run = false
    
  case "who":
    for _, pp := range pc.PlayerChars {
      fmt.Printf("%q: %s\n", pp.Ref(), pp.Full(0))
    }
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
  if DEBUG {
    dtalog.Start(dtalog.DBG, os.Stdout)
  }
  
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
  commandChannel = make(chan string, commandQueueLength)
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
  err := os.Remove(sockName)
  if err != nil {
    log(dtalog.ERR, "main(): unable to remove control socket %q: %s\n",
                    sockName, err)
  }
}

