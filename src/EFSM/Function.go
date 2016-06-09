package EFSM

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Function struct {
	Name           string
	Transitions    []*Transition
	Variable       *Variable
	apiUrl         string
	apiContentType string
	apiBody        string
	apiMethod      string
}

type FunctionJSON struct {
	Name        string           `json:"name"`
	Transitions []TransitionJSON `json:"transitions"`
	Variable    string           `json:"variable,omitempty"`
}

func (function *Function) Serialize() FunctionJSON {
	var transitions []TransitionJSON
	for i := range function.Transitions {
		transitions = append(transitions, function.Transitions[i].Serialize())
	}
	var varName string
	if function.Variable != nil {
		varName = function.Variable.Name
	}
	return FunctionJSON{Name: function.Name, Variable: varName, Transitions: transitions}
}

func (function *Function) findReplaceVariableInApiBody() string {
	if function.Variable == nil {
		return function.apiBody
	}

	var replaceValue string
	switch function.Variable.varType.String() {
	case "bool":
		replaceValue = function.Variable.Value
	case "float64":
		replaceValue = function.Variable.Value
	case "string":
		replaceValue = "\"" + function.Variable.Value + "\""
	}
	return strings.Replace(function.apiBody, "\"$var\"", replaceValue, -1)
}

func (function *Function) execute(currentState *State, arg string) (*State, error) {
	for i := range function.Transitions {
		if function.Transitions[i].From == currentState {
			if len(arg) > 0 && function.Variable != nil {
				function.Variable.setValue(arg)
			}
			if len(function.apiUrl) > 0 {
				switch function.apiContentType {
				case "JSON":
					client := &http.Client{}
					req, err := http.NewRequest(function.apiMethod, function.apiUrl, strings.NewReader(function.findReplaceVariableInApiBody()))
					if err != nil {
						return nil, err
					}
					res, err := client.Do(req)
					if err != nil {
						return nil, err
					}
					defer res.Body.Close()
					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						return nil, err
					}
					if res.StatusCode < 200 || res.StatusCode >= 300 {
						fmt.Println("Non 2xx API response: ", body)
					}
				}
			}
			return function.Transitions[i].To, nil
		}
	}
	return nil, fmt.Errorf("Function \"%s\": function called from non-linked state: %s", function.Name, currentState.Name)
}
