package EFSM

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"fmt"
)

type root struct {
	Info      infoObject
	Functions []functionObject
	Api				apiObject
}

type apiObject struct {
	ApiBase string
	ApiStateCalls []apiStateObject
}

type apiStateObject struct {
	ApiPath string
	Variables map[string]string
	States map[string]stateExpression
	Interval int
}

type stateExpression struct {
	Field string
	Operator string
	Value interface{}
}

type infoObject struct {
	Title   string
	Version string
}

type functionObject struct {
	Name        string
	Transitions []transitionObject
	Variable    string
	ApiPath			string
	ApiContentType	string
	ApiBody 	string
	ApiMethod string
}

type transitionObject struct {
	From string
	To   string
}

func GetGenericJSONMap(bytes []byte) (map[string]interface{}, error){
	generic := make(map[string]interface{})
	err := json.Unmarshal(bytes, &generic)
	if err != nil {
		return nil, err
	}
	return generic, nil
}

// TODO: Splitting and then joining again is not very efficient, could replace with
// inner recursive function that works directly on the array of paths
func GetValueFromGenericJSONMap(genJson map[string]interface{}, path string) (interface{}, error){
	paths := strings.Split(path, ".")
	if genJson[paths[0]] == nil {
		return nil, fmt.Errorf("Component %s from path %s not found in JSON %v\n", paths[0], path, genJson)
	} else if len(paths) == 1 {
		return genJson[paths[0]], nil
	} else {
		return GetValueFromGenericJSONMap(genJson[paths[0]].(map[string]interface{}), strings.Join(paths[1:], "."))
	}
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

	sr := &StateRetriever{}
	stateCalls := r.Api.ApiStateCalls
	for i := range stateCalls {
		src := NewStateRetrieveCall((r.Api.ApiBase + stateCalls[i].ApiPath), stateCalls[i].Interval)
		for key, value := range stateCalls[i].Variables {
			src.variables[efsm.addVariable(key)] = value
		}
		for key, value := range stateCalls[i].States {
			state, err := efsm.addState(key)
			if (err != nil){
				return nil, err
			}
			src.stateExpressions[state] = value
		}
		sr.states = append(sr.states, src)
	}
	efsm.stateRetriever = sr

	for i := range r.Functions {
		var fo *functionObject = &r.Functions[i]
		temp := new(Function)
		temp.name = fo.Name
		temp.apiUrl	= r.Api.ApiBase + fo.ApiPath
		temp.apiContentType	= fo.ApiContentType
		temp.apiBody = fo.ApiBody
		temp.apiMethod = fo.ApiMethod

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
	return efsm, nil
}
