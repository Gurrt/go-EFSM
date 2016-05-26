package EFSM

import "fmt"

type Function struct {
  name string
  transitions []*Transition
  variable *Variable
}

func (function *Function) execute(currentState *State, arg string) (*State, error) {
  for i := range function.transitions {
    if(function.transitions[i].from == currentState){
      if(len(arg) > 0 && function.variable != nil){
        function.variable.setValue(arg)
      }
      return function.transitions[i].to, nil
    }
  }
    return nil, fmt.Errorf("Function \"%s\": function called from non-linked state: %s", function.name, currentState.name)
}
