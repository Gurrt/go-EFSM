package EFSM

type Transition struct {
	From *State
	To   *State
}

type TransitionJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (transition *Transition) Serialize() TransitionJSON {
	return TransitionJSON{From: transition.From.Name, To: transition.To.Name}
}

func (transition *Transition) toString() string {
	return transition.To.toString() + " -> " + transition.From.toString()
}
