// chedit.go
//
// edit player character files for dta5
//
// (Formerly dta5wpd.go; the only thing it did was change passwords.)
//
// updated 2017-08-14
//
package main

import( "encoding/json"; "flag"; "fmt"; "os"; "path/filepath";
        "golang.org/x/crypto/bcrypt";
        "github.com/d2718/dconfig";
)

var pcDir string = "pc_dir"
var hashCost int = 4

func init() {
  dconfig.Reset()
  dconfig.AddInt(&hashCost, "hash_cost", dconfig.UNSIGNED)
  dconfig.Configure([]string{"conf"}, false)
}

func die(err error, fmtstr string, args ...interface{}) {
  if err != nil {
    fmt.Printf(fmtstr, args...)
    panic(err)
  }
}

type Data map[string]interface{}

func setPassword(pcDat Data, newPwd string) Data {
  new_pwd_bs, err := bcrypt.GenerateFromPassword([]byte(newPwd), hashCost)
  die(err, "Error generating new password: %s\n", err)
  
  pcDat["PassHash"] = string(new_pwd_bs)
  return pcDat
}

func resetInventory(pcDat Data) Data {
  pcDat["RightHand"] = ""
  pcDat["LeftHand"] = ""
  pcDat["Inventory"] = make([]string, 0, 0)
  return pcDat
}

var newPassword string
var reset bool

func main() {
  flag.StringVar(&newPassword, "p", "",    "specify a new password")
  flag.BoolVar(&reset,         "r", false, "reset inventory")
  flag.Parse()
  uname := flag.Arg(0)
  
  //fmt.Println(newPassword, reset)
  
  // read data
  
  fname := filepath.Join(pcDir, uname + ".json")
  f, err := os.Open(fname)
  die(err, "Unable to open %q: %s\n", fname, err)
  dcdr := json.NewDecoder(f)
  
  var raw_dat interface{}
  err = dcdr.Decode(&raw_dat)
  die(err, "Unable to read file %q: %s\n", fname, err)
  
  var other_stuff = make([]interface{}, 0, 0)
  for dcdr.More() {
    var itm interface{}
    err = dcdr.Decode(&itm)
    die(err, "Error decoding item: %s\n", err)
    other_stuff = append(other_stuff, itm)
  }
  f.Close()
  map_dat := Data(raw_dat.(map[string]interface{}))
  
  //fmt.Println(map_dat)
  
  // process data
  
  if newPassword != "" {
    map_dat = setPassword(map_dat, newPassword)
  }
  if reset == true {
    map_dat = resetInventory(map_dat)
  }
  
  //fmt.Println(map_dat)
  
  // write data
  
  f, err = os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC, 0644)
  die(err, "Error opening %q for writing: %s\n", fname, err)
  defer f.Close()
  ncdr := json.NewEncoder(f)
  
  err = ncdr.Encode(map_dat)
  die(err, "Error writing PC map %v to %q: %s\n", map_dat, fname, err)
  
  if reset == false {
    for _, x := range other_stuff {
      err = ncdr.Encode(x)
      die(err, "Error writing %v to %q: %s\n", x, fname, err)
    }
  }
}
