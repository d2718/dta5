// mood.go
//
// dta5 mood messaging
//
// updated 2017-08-10
//
package mood

import( "fmt"; "math/rand"; "time";
        "dta5/act"; "dta5/log"; "dta5/msg"; "dta5/ref"; "dta5/room";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("mood: " + fmtstr, args...))
}

type MoodMessenger struct {
  MinDelay   time.Duration
  DelayRange time.Duration
  Coverage   []string
  Messages   []string
}

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

var Messengers []*MoodMessenger

func Initialize() {
  Messengers = make([]*MoodMessenger, 0, 0)
}

func NewMessenger(min, max float64,
                  coverage, text []string) *MoodMessenger {
  
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

func (mmp *MoodMessenger) Message() {
  m := msg.New("txt", mmp.Messages[randSource.Intn(len(mmp.Messages))])
  for _, r_id := range mmp.Coverage {
    ref.Deref(r_id).(*room.Room).Deliver(m)
  }
}
