// log.go
//
// The dta5 logging package.
//
// updated 2017-08-30
//
// The dta5/dtalog package provides a centralized logging mechanism for the
// development and operation of the dta5 project.
//
// Recommended use is to have each package that needs logging import this
// one, and define its own unexported log() function
//
//  func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{})
//
// that prepends its own prefix to the message (to identify the package)
// and then calls dtalog.Log(), so that all logging messages can easily be
// collected in one place.
//
// This package provides for four levels of logging (and could be expanded
// for more, but there's plenty of debate about "log levels", and probably
// four is two or three too many):
//
//  * ERR: Things that are absolutely errors, either with the softwarevor with
//         the world data.
//  * WRN: Things that are maybe fishy and should be corrected, but shouldn't
//         necessarily be fatal.
//  * MSG: Information that should be logged, but isn't necessarily indicative
//         of a problem, like the game world being done loading, lost
//         connections, or failed login attempts.
//  * DBG: Messages that are useful during development, but should definitely
//         be turned off in production.
//
package dtalog

import( "fmt"; "io"; "time" )

type LogLvl byte

const(  ERR LogLvl = iota
        WRN
        MSG
        DBG
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
