// log.go
//
// The dta5 logging package.
//
package dtalog

import( "fmt"; "io"; "time" )

type LogLvl byte

const(  ERR LogLvl = iota
        WRN
        DBG
        MSG
)

var logFiles map[LogLvl]io.Writer
var TimeFmt string = "2006-01-02 15:04:05.000"

func init() {
  logFiles = make(map[LogLvl]io.Writer)
}

func Start(lvl LogLvl, targ io.Writer) {
  logFiles[lvl] = targ
}
func Stop(lvl LogLvl) {
  delete(logFiles, lvl)
}

func Log(lvl LogLvl, msg string) {
  targ, ok := logFiles[lvl]
  if (!ok) || (len(msg) < 1) {
    return
  }
  if msg[len(msg)-1] == '\n' {
    fmt.Fprintf(targ, "%s (%d) %s", time.Now().Format(TimeFmt), lvl, msg)
  } else {
    fmt.Fprintf(targ, "%s (%d) %s\n", time.Now().Format(TimeFmt), lvl, msg)
  }
}
