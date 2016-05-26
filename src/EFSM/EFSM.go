package EFSM

import (
  "fmt"
  "strings"
)

type EFSM struct {
  title string
  version string
  functions []*Function
  currentState *State
  states map[string]*State
  variableMap map[string]*Variable
}

func NewEFSM(title string, version string) *EFSM {
  variableMap := make(map[string]*Variable)
  states := make(map[string]*State)
  return &EFSM{title: title, version:version, variableMap: variableMap, states: states}
}

func (efsm *EFSM) ExecuteFunction(name string){
  funcArr := strings.Split(name, " ")
  var newState *State
  var err error
  if(len(funcArr) > 2){
    fmt.Println("Error: Function can have a maximum of one argument")
    return
  } else if (len(funcArr) == 2) {
    newState, err = efsm.currentState.executeFunction(funcArr[0], funcArr[1])
  } else {
    newState, err = efsm.currentState.executeFunction(funcArr[0], "")
  }
  if(err != nil){
    fmt.Println(err,"\n")
  } else {
    efsm.currentState = newState
  }
}

func (efsm *EFSM) GetCurrentStateName() string {
  return efsm.currentState.name
}

func (efsm *EFSM) Print(){
  fmt.Printf("Title: %s\nVersion: %s\n", efsm.title, efsm.version)
  fmt.Printf("Variables:\n")
  for i := range efsm.variableMap {
   fmt.Printf("\t%s : %s\n",i,efsm.variableMap[i])
}
  fmt.Printf("States:\n")
  for i := range efsm.states {
    fmt.Printf("\t%s", efsm.states[i].toString())
    if(efsm.states[i] == efsm.currentState){
      fmt.Print(" --> Active")
    }
    fmt.Print("\n")
  }
  fmt.Printf("Functions:\n")
  for i := range efsm.functions {
    fmt.Printf("\t%s\n", efsm.functions[i].name)
    fmt.Printf("\t\tTransitions:\n")
    for j := range efsm.functions[i].transitions {
      fmt.Printf("\t\t\t%s\n", efsm.functions[i].transitions[j].toString())
    }
  }
}

func (efsm *EFSM) addVariable(variableName string) *Variable {
  variable, ok := efsm.variableMap[variableName]
  if (ok){
    return variable
      } else {
    variable = &Variable{name: variableName, value: ""}
    efsm.variableMap[variableName] = variable
    return variable
  }
}

func (efsm *EFSM) addState(state string) error {
  _, ok := efsm.states[state]
  if (ok){
    return fmt.Errorf("EFSM: state %s already defined", state)
  } else {
    var newState *State = newState(state)
    efsm.states[state] = newState
    if(efsm.currentState == nil){
      fmt.Println("Setting currentState to ", newState.name)
      efsm.currentState = newState
    }
    return nil
  }
}

func (efsm *EFSM) addFunction(function *Function) error {
  for i := range function.transitions {
    var state *State = function.transitions[i].from
    err := state.addFunction(function)
    if(err != nil) {
      return err
    }
  }
  efsm.functions = append(efsm.functions, function)
  return nil
}

func (efsm *EFSM) newTransition(from string, to string) (*Transition, error) {
  fromState, ok := efsm.states[from]
  if(!ok){
    return nil, fmt.Errorf("EFSM: from state %s not defined", from)
  }

  toState, ok2 := efsm.states[to]
  if(!ok2){
    return nil, fmt.Errorf("EFSM: from state %s not defined", to)
  }

  return &Transition{from: fromState, to: toState}, nil
}
