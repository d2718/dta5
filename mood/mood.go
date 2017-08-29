// mood.go
//
// dta5 mood messaging
//
// updated 2017-08-28
//
// The dta5/mood package adds atmospheric messaging and effects. These have
// no mechanical effect on the game, but hopefully contribute to the
// atmosphere.
//
package mood

import( "fmt"; "math/rand"; "time";
        "dta5/act"; "dta5/log"; "dta5/msg"; "dta5/ref"; "dta5/room";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("mood: " + fmtstr, args...))
}

// "Mood messaging" is text delivered to the occupants of a set of rooms in
// order to enhance the feel of the location. A MoodMessenger delivers a
// message randomly chosen from a specified list to its set of room.Rooms on
// a semi-periodic basis.
//
type MoodMessenger struct {
  MinDelay   time.Duration
  DelayRange time.Duration
  Coverage   []string
  Messages   []string
}

// This once source of randomness serves the whole package.
var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// MoodMessengers generally don't require or warrant any kind of "interaction"
// during the course of the game, but, once loaded, need somewhere to "be" to
// avoid garbage collection. That place is here.
//
var Messengers []*MoodMessenger

// This function prepares the package for loading the game. This should be
// called both on initial game loading and when loading a saved game state.
//
func Initialize() {
  Messengers = make([]*MoodMessenger, 0, 0)
}

// This function is meant to be called by the dta5/load package to load
// mood messaging objects. min and max are the minimum and maximum delays
// between message delivery (the time is randomized between these two
// values each delivery cycle); coverage is a slice of room reference IDs;
// text is the list of messages that could be delivered (a random one is
// chosen every delivery cycle).
//
func NewMessenger(min, max float64, coverage, text []string) *MoodMessenger {
  
  rng := max - min
  nmmp := &MoodMessenger{
    MinDelay:   time.Duration(min * 1000000000),
    DelayRange: time.Duration(rng * 1000000000),
    Coverage:   coverage,
    Messages:   text,
  }
  Messengers = append(Messengers, nmmp)
  
  log(dtalog.DBG, "*MoodMessenger(%f, %f, %q, %d msgs) added",
      nmmp.MinDelay.Seconds(), nmmp.DelayRange.Seconds(), nmmp.Coverage,
      len(nmmp.Messages))
  
  return nmmp
}

// When a MoodMessenger is loaded by dta5/load, this function gets it started
// delivering messages (by sticking itself in the dta5/act ActionQueue).
//
func (mmp *MoodMessenger) Arm() {
  delay := mmp.MinDelay + time.Duration(randSource.Int63n(int64(mmp.DelayRange)))
  f := func() error {
    mmp.Message()
    mmp.Arm()
    return nil
  }
  a := act.Action{
    Time: time.Now().Add(delay),
    Act: f,
  }
  act.Enqueue(&a)
}

// This is the function that chooses and delivers the message. It is called
// by enqued act.Action.
//
func (mmp *MoodMessenger) Message() {
  m := msg.New("txt", mmp.Messages[randSource.Intn(len(mmp.Messages))])
  for _, r_id := range mmp.Coverage {
    ref.Deref(r_id).(*room.Room).Deliver(m)
  }
}
