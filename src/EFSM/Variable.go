package EFSM

import (
	"reflect"
	"sync"
)

type VariableJSON struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Variable struct {
	sync.Mutex
	Name    string
	Value   string
	varType reflect.Type
}

func (variable *Variable) Serialize() VariableJSON {
	return VariableJSON{Name: variable.Name, Value: variable.Value}
}

func (variable *Variable) setValue(newValue string) {
	variable.Lock()
	variable.Value = newValue
	variable.Unlock()
}
