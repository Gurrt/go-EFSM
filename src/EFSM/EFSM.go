package EFSM

import (
	"fmt"
	"strings"
)

type EFSM struct {
	ID             string
	stateRetriever *StateRetriever
	functions      []*Function
	currentState   *State
	states         map[string]*State
	variableMap    map[string]*Variable
	c              chan *State
}

func NewEFSM(id string) *EFSM {
	variableMap := make(map[string]*Variable)
	states := make(map[string]*State)
	c := make(chan *State)
	return &EFSM{ID: id, variableMap: variableMap, states: states, c: c}
}

func (efsm *EFSM) Init() {
	efsm.stateRetriever.init(efsm.c)
	c := make(chan int)
	go efsm.stateUpdateListener(c)
	// Only return after we got the first stateUpdate
	_ = <-c
}

func (efsm *EFSM) Kill() {
	fmt.Println("TODO: Implement code so EFSM can be killed safely")
}

func (efsm *EFSM) stateUpdateListener(c chan int) {
	var initial = true
	for state := range efsm.c {
		if efsm.currentState != state {
			efsm.currentState = state
			fmt.Println("[", efsm.ID, "] Updated current state to :", state.name)
			if initial {
				initial = false
				c <- 1
				close(c)
			}
		}
	}
}

func (efsm *EFSM) ExecuteFunction(name string) error {
	funcArr := strings.Split(name, " ")
	var newState *State
	var err error
	if len(funcArr) > 2 {
		return fmt.Errorf("Error: Function can have a maximum of one argument")
	} else if len(funcArr) == 2 {
		newState, err = efsm.currentState.executeFunction(funcArr[0], funcArr[1])
	} else {
		newState, err = efsm.currentState.executeFunction(funcArr[0], "")
	}
	if err != nil {
		return err
	} else {
		efsm.currentState = newState
	}
	return nil
}

func (efsm *EFSM) Print() {
	fmt.Printf("Id: %s\n", efsm.ID)
	fmt.Printf("Variables:\n")
	for i := range efsm.variableMap {
		fmt.Printf("\t%s : %s\n", i, efsm.variableMap[i].value)
	}
}

func (efsm *EFSM) addVariable(variableName string) *Variable {
	variable, ok := efsm.variableMap[variableName]
	if ok {
		return variable
	} else {
		variable = &Variable{name: variableName, value: ""}
		efsm.variableMap[variableName] = variable
		return variable
	}
}

func (efsm *EFSM) addState(state string) (*State, error) {
	_, ok := efsm.states[state]
	if ok {
		return nil, fmt.Errorf("EFSM: state %s already defined", state)
	} else {
		var newState *State = newState(state)
		efsm.states[state] = newState
		return newState, nil
	}
}

func (efsm *EFSM) addFunction(function *Function) error {
	for i := range function.transitions {
		var state *State = function.transitions[i].from
		err := state.addFunction(function)
		if err != nil {
			return err
		}
	}
	efsm.functions = append(efsm.functions, function)
	return nil
}

func (efsm *EFSM) newTransition(from string, to string) (*Transition, error) {
	fromState, ok := efsm.states[from]
	if !ok {
		return nil, fmt.Errorf("EFSM: from state %s not defined", from)
	}

	toState, ok2 := efsm.states[to]
	if !ok2 {
		return nil, fmt.Errorf("EFSM: from state %s not defined", to)
	}

	return &Transition{from: fromState, to: toState}, nil
}
