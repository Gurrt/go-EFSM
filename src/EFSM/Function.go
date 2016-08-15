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
	EFSMid         string
}

type FunctionJSON struct {
	Name        string           `json:"name"`
	Transitions []TransitionJSON `json:"transitions"`
	Variable    string           `json:"variable,omitempty"`
	URL         string           `json:"url"`
}

func (function *Function) Serialize(baseURL string) FunctionJSON {
	var transitions []TransitionJSON
	for i := range function.Transitions {
		transitions = append(transitions, function.Transitions[i].Serialize())
	}
	var varName string
	if function.Variable != nil {
		varName = function.Variable.Name
	}
	return FunctionJSON{Name: function.Name, Variable: varName, Transitions: transitions, URL: fmt.Sprintf("%s/%s", baseURL, function.Name)}
}

func (function *Function) findReplaceVariablesInApiBody() string {
	result := function.apiBody
	if function.Variable != nil {
		var replaceValue string = function.Variable.Value
		if function.Variable.VarType == "string" {
			replaceValue = "\"" + function.Variable.Value + "\""
		}
		result = strings.Replace(result, "\"$var\"", replaceValue, -1)
	}
	result = strings.Replace(result, "$id", function.EFSMid, -1)
	return result
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
					fmt.Printf("Sending %s to %s\n", strings.NewReader(function.findReplaceVariablesInApiBody()), function.apiUrl)
					req, err := http.NewRequest(function.apiMethod, function.apiUrl, strings.NewReader(function.findReplaceVariablesInApiBody()))
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
						fmt.Println("Non 2xx API response: ", string(body))
					} else {
						fmt.Println("Successful API response: ", string(body))
					}
				}
			}
			return function.Transitions[i].To, nil
		}
	}
	return nil, fmt.Errorf("Function \"%s\": function called from non-linked state: %s", function.Name, currentState.Name)
}
