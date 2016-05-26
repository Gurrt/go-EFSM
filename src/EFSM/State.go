package EFSM

import "fmt"

type State struct {
  name string
  functions map[string]*Function
}

func newState(name string) *State {
  functions := make(map[string]*Function)
  return &State{name: name, functions: functions}
}

func (state *State) executeFunction(name string, args string) (*State, error) {
  function, ok := state.functions[name]
  if(!ok){
    return nil, fmt.Errorf("State \"%s\": function %s does not exist or is not callable from this state. Callable functions: %v", state.name, name, state.functions)
  } else {
    newState, err := function.execute(state, args)
    if(err != nil){
      return nil, err
    } else {
      return newState, nil
    }
  }
}

func (state *State) addFunction(function *Function) error {
  _, ok := state.functions[function.name]
  if (ok){
    return fmt.Errorf("State: function %s already defined", function.name)
  } else {
    state.functions[function.name] = function
    return nil
  }
}

func (state *State) toString() string {
  return state.name
}
