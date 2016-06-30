package EFSM

import (
	"sync"
)

type VariableJSON struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
	Value   string `json:"value"`
	Type    string `json:"type"`
}

type Variable struct {
	sync.Mutex
	Name    string
	Profile string
	VarType string
	Value   string
}

func (variable *Variable) Serialize() VariableJSON {
	return VariableJSON{Name: variable.Name, Profile: variable.Profile, Value: variable.Value, Type: variable.VarType}
}

func (variable *Variable) setValue(newValue string) {
	variable.Lock()
	variable.Value = newValue
	variable.Unlock()
}
