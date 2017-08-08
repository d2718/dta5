// help.go
//
// dta5 help system
//
// updated 2017-08-06
//
package pc

import( "bytes"; "os"; "path/filepath"; "sort"; "strings";
        "dta5/log"; "dta5/thing";
)

var HelpDir string

const moreHelp = `
Additional help is available by typing

> %s <option>

where <option> is one of the following:

%s`

func DoHelp(pp *PlayerChar, verb string, dobj thing.Thing,
            prep string, iobj thing.Thing, text string) {
  
  raw_toks := strings.Fields(text)
  toks := make([]string, 0, len(raw_toks))
  uc_toks := make([]string, 0, len(raw_toks))
  for _, x := range raw_toks {
    toks = append(toks, strings.ToLower(x))
    uc_toks = append(uc_toks, strings.ToUpper(x))
  }
  toks[0] = HelpDir
  
  pth := filepath.Join(toks...)
  pth_fi, err := os.Stat(pth)
  if err != nil {
    pp.QWrite("There is no help for topic %s.", strings.Join(uc_toks[1:], " "))
    return
  }
  var addl_opts []string = make([]string, 0, 0)
  if pth_fi.Mode().IsDir() {
    dir_f, err := os.Open(pth)
    if err != nil {
      log(dtalog.ERR, "DoHelp(): error opening directory %q for reading: %s",
                      pth, err)
    } else {
      fnames, err := dir_f.Readdirnames(0)
      if err != nil {
        log(dtalog.ERR, "DoHelp(): error reading directory %q: %s", pth, err)
      } else {
        for _, x := range fnames {
          if x != "_" {
            addl_opts = append(addl_opts, strings.ToUpper(x))
          }
        }
      }
    }
    pth = filepath.Join(pth, "_")
  }
  
  f, err := os.Open(pth)
  if err != nil {
    log(dtalog.ERR, "DoHelp(): unable to open %q: %s", pth, err)
    pp.QWrite("There was an error trying to display the help for topic %s.",
              strings.Join(uc_toks[1:], " "))
    return
  }
  defer f.Close()
  
  buff := new(bytes.Buffer)
  _, err = buff.ReadFrom(f)
  if err != nil {
    log(dtalog.ERR, "DoHelp(): error reading from %q: %s", pth, err)
    pp.QWrite("There was an error trying to display the help for topic %s.",
              strings.Join(uc_toks[1:], " "))
    return
  }
  
  pp.QWrite("\n%s\n", strings.Join(uc_toks, " "))
  pp.QWrite(strings.TrimSpace(buff.String()))
  if len(addl_opts) > 0 {
    sort.Sort(sort.StringSlice(addl_opts))
    pp.QWrite(moreHelp, strings.Join(uc_toks, " "), strings.Join(addl_opts, " "))
  }
}
