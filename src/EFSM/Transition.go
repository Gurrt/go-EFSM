package EFSM

type Transition struct {
  from *State
  to *State
}

func (transition *Transition) toString() string {
  return transition.to.toString() + " -> " + transition.from.toString()
}
