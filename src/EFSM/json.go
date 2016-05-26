package EFSM

import (
	"encoding/json"
	"io/ioutil"
)

type root struct {
	Info      infoObject
	States    []string
	Functions []functionObject
}

type infoObject struct {
	Title   string
	Version string
}

type functionObject struct {
	Name        string
	Transitions []transitionObject
	Variable    string
}

type transitionObject struct {
	From string
	To   string
}

func FromJSONFile(filename string) (*EFSM, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var r root
	err = json.Unmarshal(contents, &r)
	if err != nil {
		return nil, err
	}
	efsm := NewEFSM(r.Info.Title, r.Info.Version)

	for i := range r.States {
		efsm.addState(r.States[i])
	}

	for i := range r.Functions {
		var fo *functionObject = &r.Functions[i]
		temp := new(Function)
		temp.name = fo.Name
		if fo.Variable != "" {
			temp.variable = efsm.addVariable(fo.Variable)
		}
		for j := range fo.Transitions {
			var t transitionObject = fo.Transitions[j]
			trans, err := efsm.newTransition(t.From, t.To)
			if err != nil {
				return nil, err
			}
			temp.transitions = append(temp.transitions, trans)
		}
		err = efsm.addFunction(temp)
		if err != nil {
			return nil, err
		}
	}
	efsm.Print()
	return efsm, nil
}
