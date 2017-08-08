// util.go
//
// dta5 string utilities
//
package util

import( "fmt"; "strings"
)

var decapitalizer rune = 'a' - 'A'

func Cap(s string) string {
  if len(s) > 0 {
    runez := []rune(s)
    if (runez[0] >= 'a') && (runez[0] <= 'z') {
      runez[0] -= decapitalizer
      return string(runez)
    } else {
      return s
    }
  } else {
  return ""
  }
}

func EnglishList(stuff []string) string {
  switch len(stuff) {
  case 0:
    return ""
  case 1:
    return stuff[0]
  case 2:
    return fmt.Sprintf("%s and %s", stuff[0], stuff[1])
  default:
    last := stuff[len(stuff)-1]
    rest := stuff[:len(stuff)-1]
    return fmt.Sprintf("%s, and %s", strings.Join(rest, ", "), last)
  }
}
