// util.go
//
// dta5 string utilities
//
// updated 2017-08-29
//
// Package dta5/util provides some miscellaneous (mainly string manipulation)
// utilities that will be useful in several packages.
//
package util

import( "fmt"; "strings" )

// decapitalizer is the value that must be subtracted from an ASCII-range
// rune representing a lower-case letter in order to make it upper-case.
//
var decapitalizer rune = 'a' - 'A'

// Cap() returns its argument with the first character capitalized (or
// unchanged if it's already capitalized or starts with a character for which
// capitalization isn't meaningful).
//
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

// EnglishList() returns the strings in its argument slice as a single
// string in an appropriately comma-separated, English-language list.
//
//  EnglishList([]string{}) => ""
//  EnglishList([]string{"a mother"}) => "a mother"
//  EnglishList([]string{"a mother", "a maiden"})
//      => "a mother and a maiden"
//  EnglishList([]string{"a mother", "a maiden", "a crone"})
//      => "a mother, a maiden, and a crone"
//  EnglishList([]string{"a mother", "a maiden", "a crone", "an enchantment"})
//      => "a mother, a maiden, a crone, and an enchantment"
//
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
