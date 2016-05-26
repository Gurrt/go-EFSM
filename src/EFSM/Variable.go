package EFSM

type Variable struct {
  name string
  value string
}

func (variable *Variable) setValue(newValue string){
  variable.value = newValue
}
