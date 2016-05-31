package EFSM

import (
  "reflect"
  "sync"
)

type Variable struct {
  sync.Mutex
  name string
  value string
  varType reflect.Type
}

func (variable *Variable) setValue(newValue string){
  variable.Lock()
  variable.value = newValue
  variable.Unlock()
}
