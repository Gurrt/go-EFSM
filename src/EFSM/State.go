package EFSM

import "fmt"

type State struct {
	Name      string
	Functions map[string]*Function
}

func newState(name string) *State {
	functions := make(map[string]*Function)
	return &State{Name: name, Functions: functions}
}

func (state *State) executeFunction(name string, args string) (*State, error) {
	function, ok := state.Functions[name]
	if !ok {
		return nil, fmt.Errorf("State \"%s\": function %s does not exist or is not callable from this state. Callable functions: %v", state.Name, name, state.Functions)
	} else {
		newState, err := function.execute(state, args)
		if err != nil {
			return nil, err
		} else {
			return newState, nil
		}
	}
}

func (state *State) addFunction(function *Function) error {
	_, ok := state.Functions[function.Name]
	if ok {
		return fmt.Errorf("State: function %s already defined", function.Name)
	} else {
		state.Functions[function.Name] = function
		return nil
	}
}

func (state *State) toString() string {
	return state.Name
}
