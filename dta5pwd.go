// pwd.go
//
// generate password hashes for dta5
//
// updated 2017-08-07
//
package main

import( "encoding/json"; "flag"; "fmt"; "os"; "path/filepath";
        "golang.org/x/crypto/bcrypt";
        "github.com/d2718/dconfig";
)

var pcDir string = "pc_dir"
var hashCost int = 4

func configure() {
  dconfig.Reset()
  dconfig.AddInt(&hashCost, "hash_cost", dconfig.UNSIGNED)
  dconfig.Configure([]string{"conf"}, true)
}

func main() {
  flag.Parse()
  uname := flag.Arg(0)
  new_pwd := flag.Arg(1)
  
  fname := filepath.Join(pcDir, uname + ".json")
  f, err := os.Open(fname)
  if err != nil {
    fmt.Printf("Unable to open %q: %s\n", fname, err)
    os.Exit(1)
  }
  dcdr := json.NewDecoder(f)
  
  var raw_dat interface{}
  err = dcdr.Decode(&raw_dat)
  if err != nil {
    fmt.Printf("Unable to read %q: %s\n", fname, err)
    os.Exit(1)
  }
  
  var other_stuff = make([]interface{}, 0, 0)
  for dcdr.More() {
    var itm interface{}
    err = dcdr.Decode(&itm)
    if err != nil {
      fmt.Printf("Error decoding item: %s\n", err)
      os.Exit(1)
    }
    other_stuff = append(other_stuff, itm)
  }
  f.Close()
  map_dat := raw_dat.(map[string]interface{})
  
  new_pwd_bs, err := bcrypt.GenerateFromPassword([]byte(new_pwd), hashCost)
  if err != nil {
    fmt.Printf("Error generating new password: %s\n", err)
    os.Exit(1)
  }
  
  map_dat["PassHash"] = string(new_pwd_bs)
  
  f, err = os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC, 0644)
  if err != nil {
    fmt.Printf("Error opening %q for writing: %s\n", fname, err)
    os.Exit(1)
  }
  defer f.Close()
  ncdr := json.NewEncoder(f)
  
  err = ncdr.Encode(map_dat)
  if err != nil {
    fmt.Printf("Error writing PC map %v to %q: %s\n", map_dat, fname, err)
    os.Exit(1)
  }
  
  for _, x := range other_stuff {
    err = ncdr.Encode(x)
    if err != nil {
      fmt.Printf("Error writing %v to %q: %s\n", x, fname, err)
      os.Exit(1)
    }
  }
}
