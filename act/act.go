// act.go
//
// dta5 action queue
//
// 2017-08-28
//
// The game involves a single queue of *Action pointers, which are evaluated
// sequentially and at the appropriate times in a semi-real-time busy-wait loop.
//
// An Action represents anything that should "happen" in the game, including
// delivering mood messaging, parsing (and acting upon) player commands, and
// the AIs of creatures deciding what to do (and doing it) (although as of
// the writing of this comment --2017-08-28-- the latter has not yet been
// implemented). An Action contains two elements: a time.Time at which the
// action should happen, and a function taking no arguments and returning
// an error that should be executed when the Action "happens".
//
// This package maintains a single queue of pointers to Actions as a heap.
// Two functions will add *Actions to the heap: Enqueue() and Add() (see below
// for details on each), and one will remove them: Next(). Next() returns
// an action if the Time of the latest Action in the queue has passed, or nil
// otherwise.
//
package act

import( "container/heap"; "fmt"; "time";
        "dta5/log";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("act: " + fmtstr, args...))
}
const logTimeFmt = "15:04:05.000000"

// An Action represents a single "thing" that should "happen". They are
// ordered by when they should happen.
//
type Action struct {
  time.Time
  Act func() error
}

// This package maintains a single one of these as a heap, and has a single
// function for adding and popping them. The actual calls to heap.Push() and
// heap.Pop() occur in a single goroutine (which sends and receives them
// via channels, so it's thread-safe.
//
type ActionQueue []*Action

// ActionQueue implements sort.Interface

func (aq ActionQueue) Len() int { return len(aq) }
func (aq ActionQueue) Swap(i, j int) { aq[i], aq[j] = aq[j], aq[i] }
func (aq ActionQueue) Less(i, j int) bool { return aq[j].After(aq[i].Time) }

// ActionQueue also implements heap.Interface

func (aqp *ActionQueue) Push(x interface{}) {
  var ap *Action
  
  switch a := x.(type) {
  case Action:
    ap = &a
  case *Action:
    ap = a
  }
  
  (*aqp) = append(*aqp, ap)
}

func (aqp *ActionQueue) Pop() interface{} {
  n := len(*aqp)
  x := (*aqp)[n-1]
  (*aqp) = (*aqp)[0:n-1]
  return x
}

func (aqp *ActionQueue) More() bool {
  if aqp.Len() > 0 {
    if time.Now().After((*aqp)[0].Time) {
      return true
    }
  }
  return false
}

// This is the single queue of *Actions maintained by this package, and the
// channels used by the idle() goroutine to send and receive them.
var q *ActionQueue
var pushChan chan *Action
var popChan  chan *Action

// This is the length of the busy-wait interval idle() uses to check the
// queue for Actions that are due. It may warrant tuning at some later time.
//
var IdleInterval int = 100 //milliseconds

// Starts or the management of the single queue of *Actions. Only call this
// once, because it starts the idle() goroutine that will run forever.
//
func Initialize(chanSize int) {
  new_aq := make(ActionQueue, 0, 0)
  pushChan = make(chan *Action, chanSize)
  popChan  = make(chan *Action, chanSize)
  
  q = &new_aq
  heap.Init(q)
  
  go idle()
}

// This goroutine does all of the actual pushing to and popping from the
// queue of *Actions, sending and receiving on channels as necessary, so
// this package is thread-safe.
//
func idle() {
  for {
    select {
    
    case ap := <- pushChan:
      heap.Push(q, ap)
    
    default:
      if q.More() && (len(popChan) < cap(popChan)) {
        popChan <- heap.Pop(q).(*Action)
      } else {
        time.Sleep(time.Duration(IdleInterval) * time.Millisecond)
      }
    }
  }
}

// Sticks a single *Action in the queue.
//
func Enqueue(ap *Action) {
  log(dtalog.DBG, "Enqueue(): adding *Action with Time %s", ap.Time.Format(logTimeFmt))
  pushChan <- ap
}

// Creates a new *Action and sticks it in the queue, using the supplied
// function. The delay parameter is the number of seconds after the call to
// this function that the Action should occur.
//
func Add(delay float64, f func() error) {
  log(dtalog.DBG, "Add(%f, []) called", delay)
  dly := delay * 1000000000
  a := Action{
    Time: time.Now().Add(time.Duration(dly)),
    Act: f,
  }
  Enqueue(&a)
}

// Pops the next *Action off the queue and returns it. If the queue is empty,
// or the next Action's time has not yet come, returns nil.
//
func Next() *Action {
  select {
  case ap := <- popChan:
    log(dtalog.DBG, "Next(): popping *Action with Time %s", ap.Time.Format(logTimeFmt))
    return ap
  default:
    return nil
  }
}
