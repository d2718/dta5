// act.go
//
// dta5 action queue
//
// 2017-07-24
//
package act

import( "container/heap"; "fmt"; "time";
        "dta5/log";
)

func log(lvl dtalog.LogLvl, fmtstr string, args ...interface{}) {
  dtalog.Log(lvl, fmt.Sprintf("act.go: " + fmtstr, args...))
}
const logTimeFmt = "15:04:05.000000"

type Action struct {
  time.Time
  Act func() error
}

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

var q *ActionQueue
var pushChan chan *Action
var popChan  chan *Action

var IdleInterval int = 100 //milliseconds

func Initialize(chanSize int) {
  new_aq := make(ActionQueue, 0, 0)
  pushChan = make(chan *Action, chanSize)
  popChan  = make(chan *Action, chanSize)
  
  q = &new_aq
  heap.Init(q)
  
  go idle()
}

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

func Enqueue(ap *Action) {
  log(dtalog.DBG, "Enqueue(): adding *Action with Time %s", ap.Time.Format(logTimeFmt))
  pushChan <- ap
}

func Add(delay float64, f func() error) {
  log(dtalog.DBG, "Add(%f, []) called", delay)
  dly := delay * 1000000000
  a := Action{
    Time: time.Now().Add(time.Duration(dly)),
    Act: f,
  }
  Enqueue(&a)
}
    

func Next() *Action {
  select {
  case ap := <- popChan:
    log(dtalog.DBG, "Next(): popping *Action with Time %s", ap.Time.Format(logTimeFmt))
    return ap
  default:
    return nil
  }
}
