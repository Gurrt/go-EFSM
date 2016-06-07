package EFSM

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Function struct {
	name           string
	transitions    []*Transition
	variable       *Variable
	apiUrl         string
	apiContentType string
	apiBody        string
	apiMethod      string
}

func (function *Function) findReplaceVariableInApiBody() string {
	if function.variable == nil {
		return function.apiBody
	}

	var replaceValue string
	switch function.variable.varType.String() {
	case "bool":
		replaceValue = function.variable.value
	case "float64":
		replaceValue = function.variable.value
	case "string":
		replaceValue = "\"" + function.variable.value + "\""
	}
	return strings.Replace(function.apiBody, "\"$var\"", replaceValue, -1)
}

func (function *Function) execute(currentState *State, arg string) (*State, error) {
	for i := range function.transitions {
		if function.transitions[i].from == currentState {
			if len(arg) > 0 && function.variable != nil {
				function.variable.setValue(arg)
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
			return function.transitions[i].to, nil
		}
	}
	return nil, fmt.Errorf("Function \"%s\": function called from non-linked state: %s", function.name, currentState.name)
}
