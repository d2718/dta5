// scripts.go
//
// The dta5/scripts package provides a framework for specifying per-object
// custom behavior.
//
// updated 2017-08-29
//
// Each thing.Thing behaves in a predictable way to being the object of any
// given verb, eg., a thing.Container that IsToggleable() will open or close
// when a PC attempts to OPEN or CLOSE it; a thing.Item of sufficently low
// weight and bulk will be picked up when someone with a free hand tries to
// GET it; an open door.Doorway will take someone elsewhere when he tries
// to GO through it, etc.
//
// This package provides a framework for, eg., making a container that can
// only be unlocked with a specific key; an item that denies being picked up
// using custom messaging; a portal that can't be traversed unless the person
// going through it is wearing a specific amulet, etc.
//
// The mechanism for providing this custom behavior for an object involves
// associating a particular (verb, function) pair with a thing.Thing. Whenever
// a particular Thing is to be the object of an action, a check is made to
// see if there is such an association involving the verb of the action to
// be performed. If extant, the associated function is called, either
// superseding or augmenting the default behavior as specified.
//
// The functions specifying custom behavior must all have the same signature
// (see the Script type, below), but can be defined anywhere that imports
// this package. By default, there is a dta5/scripts/more package where I have
// collected the majority of the custom scripts that I have written, but you
// can put them anywhere.
//
package scripts

import( "fmt";
        "dta5/log"; "dta5/ref"; "dta5/save"; "dta5/thing";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("scripts: " + fmtstr, args...))
}

// Log is meant to be used by Scripts (defined in any package) to report
// anything unexpected or fishy. They are all reported at the level of
// "warning" (that is, dtalog.WRN), and one should take great pains when
// writing a script to ensure that any possible errors are non-fatal and
// leave the game in a reasonable, recoverable state.
//
func Log(fmtstr string, args ...interface{}) {
  dtalog.Log(dtalog.WRN, fmt.Sprintf("_script_error_: " + fmtstr, args...))
}

// All functions intended to specify default behavior should be of type Script.
// This function (with the appropriate arguments) will be called before an
// action with the associated verb is to be performed upon the associated
// object. If it returns true, the action will continue as normal; if it
// returns false, the action will not. (A result of false does not necessarily
// mean the action doesn't happen, just that all the mechanics associated
// by the action happening are handled by the Script function, and the normal
// code for performing that action doesn't need to do anything).
//
// Arguments are as follows:
//
//  obj:  the thing.Thing with which the script is associated
//  subj: the Thing (probably a pc.PlayerChar) performing the action
//  dobj: the direct object of the action (may == obj)
//  iobj: the indirect object of the action (may == obj)
//  verb: the associated verb
//  prep: the preposition specifying the relationship of iobj (if any)
//  text: the full text of command, if it's a PlayerChar action
//
type Script func(obj, subj, dobj, iobj thing.Thing,
                 verb, prep, text string) bool

// Scripts associates each Script with its identifying string.
// This allows expressing Script binding relationships in textual savegame
// and world-specification files.
//
var Scripts = make(map[string]Script)
//
// defined in pc/lock.go
//
//   * "locked_script"
//     Prevents a player from opening a thing.Openable and informs him that
//     it's locked, unless the associated data value "locked_script_unlocked"
//     is set to true.
//
//   * "lock_unlock_script"
//     Designates a thing.Thing to be a "key" that allows a thing.Openable
//     under the influence of the "locked_script" to be UNLOCKed by it.
//
// defined in pc/open.go
//
//   * "auto_close_script"
//     Causes a thing.Openable to shut automatically after it has been opened
//    (after a specified delay).

// Bindings stores the bindings between objects, verbs, and scripts.
// The keys to the map are the ref index strings of the objects to which
// scripts are bound. Each value in the map is either
//   * nil (that is, the zero value) if the object has no associated scripts
//   * another map that maps verbs with the identifying strings of the
//     scripts associated with those verbs.
// Outside of saving and loading, this map should never be accessed outside
// of using the Bind() and Check() functions (below).
//
var Bindings = make(map[string]map[string]string)

// Check() determines if an object has a script associated with a particular
// verb and calls that script if it does. This function should be called
// by _all_ code that performs "game actions" on things.
//
func Check(subj, dobj, iobj thing.Thing, verb, prep, text string) bool {
  
  if iobj != nil {
    if f, sok := Scripts[Bindings[iobj.Ref()][verb]]; sok {
      cont := f(iobj, subj, dobj, iobj, verb, prep, text)
      if !cont {
        return false
      }
    }
  }
  
  if dobj != nil {
    if f, sok := Scripts[Bindings[dobj.Ref()][verb]]; sok {
      cont := f(dobj, subj, dobj, iobj, verb, prep, text)
      if !cont {
        return false
      }
    }
  }
  
  return true
}

// Bind() associates (verb, Script) pairs with objects. The ftag parameter
// is the identifying string of the given Script function one desires to bind.
//
func Bind(obj ref.Interface, verb string, ftag string) {
  if _, ok := Scripts[ftag]; ok {
    if _, mok := Bindings[obj.Ref()]; !mok {
      Bindings[obj.Ref()] = make(map[string]string)
    }
    Bindings[obj.Ref()][verb] = ftag
  } else {
    log(dtalog.WRN, "Bind(%q, %q, %q): %q is not a script tag", obj.Ref(), verb, ftag)
  }
}

// This is called during the saving process to save all script bindings.
//
func SaveBindings(s save.Saver) {
  for t_ref, vmap := range Bindings {
    for verb, script_tag := range vmap {
      data := []interface{}{"script", t_ref, verb, script_tag}
      s.Encode(data)
    }
  }
}
