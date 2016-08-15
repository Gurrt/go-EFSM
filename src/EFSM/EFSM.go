package EFSM

import (
	"fmt"
	"sort"
)

type EFSM struct {
	ID             string
	stateRetriever *StateRetriever
	Functions      []*Function
	CurrentState   *State
	States         map[string]*State
	VariableMap    map[string]*Variable
	c              chan *State
}

type InstanceJSON struct {
	ID           string         `json:"id"`
	Variables    []VariableJSON `json:"variables"`
	CurrentState string         `json:"currentState"`
}

func (efsm *EFSM) Serialize() InstanceJSON {
	var keys []string
	var variables []VariableJSON
	for key := range efsm.VariableMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i := range keys {
		variables = append(variables, efsm.VariableMap[keys[i]].Serialize())
	}

	return InstanceJSON{ID: efsm.ID,
		Variables:    variables,
		CurrentState: efsm.CurrentState.Name}
}

func NewEFSM(id string) *EFSM {
	variableMap := make(map[string]*Variable)
	states := make(map[string]*State)
	c := make(chan *State)
	return &EFSM{ID: id, VariableMap: variableMap, States: states, c: c}
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
		if efsm.CurrentState != state {
			efsm.CurrentState = state
			fmt.Println("[", efsm.ID, "] Updated current state to :", state.Name)
			if initial {
				initial = false
				c <- 1
				close(c)
			}
		}
	}
}

func (efsm *EFSM) ExecuteFunction(name string, value string) error {
	var newState *State
	var err error
	newState, err = efsm.CurrentState.executeFunction(name, value)
	if err != nil {
		return err
	} else {
		efsm.CurrentState = newState
	}
	return nil
}

func (efsm *EFSM) Print() {
	fmt.Printf("Id: %s\n", efsm.ID)
	fmt.Printf("Variables:\n")
	for i := range efsm.VariableMap {
		fmt.Printf("\t%s : %s\n", i, efsm.VariableMap[i].Value)
	}
}

func (efsm *EFSM) setVariables(variables map[string]Variable) {
	pointerMap := make(map[string]*Variable)
	for key := range variables {
		varObj := variables[key]
		fmt.Println("Handing variable: ", key, " addr: ", &varObj)
		pointerMap[key] = &varObj
	}
	efsm.VariableMap = pointerMap
}

func (efsm *EFSM) getVariable(variable string) *Variable {
	return efsm.VariableMap[variable]
}

func (efsm *EFSM) addState(state string) (*State, error) {
	_, ok := efsm.States[state]
	if ok {
		return nil, fmt.Errorf("EFSM: state %s already defined", state)
	} else {
		newState := newState(state)
		efsm.States[state] = newState
		return newState, nil
	}
}

func (efsm *EFSM) addFunction(function *Function) error {
	for i := range function.Transitions {
		var state *State = function.Transitions[i].From
		err := state.addFunction(function)
		if err != nil {
			return err
		}
	}
	function.EFSMid = efsm.ID
	efsm.Functions = append(efsm.Functions, function)
	return nil
}

func (efsm *EFSM) newTransition(from string, to string) (*Transition, error) {
	fromState, ok := efsm.States[from]
	if !ok {
		return nil, fmt.Errorf("EFSM: from state %s not defined", from)
	}

	toState, ok2 := efsm.States[to]
	if !ok2 {
		return nil, fmt.Errorf("EFSM: from state %s not defined", to)
	}

	return &Transition{From: fromState, To: toState}, nil
}
