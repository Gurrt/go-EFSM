package EFSM

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type StateRetriever struct {
	states []*StateRetrieveCall
}

// TODO: Add support for other methods than GET
type StateRetrieveCall struct {
	variables map[*Variable]string
	url       string
	interval  int
	// stateExpression is defined in json.go
	stateExpressions map[*State]stateExpression
}

func (src *StateRetrieveCall) retrieve(c chan *State) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", src.url, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	json, err := GetGenericJSONMap(body)
	if err != nil {
		return err
	}
	for i := range src.variables {
		value, err := GetValueFromGenericJSONMap(json, src.variables[i])
		if err != nil {
			return err
		}
		if i.varType == nil {
			i.varType = reflect.TypeOf(value)
		}
		switch value.(type) {
		case string:
			i.setValue(value.(string))
		case bool:
			i.setValue(strconv.FormatBool(value.(bool)))
		case float64:
			i.setValue(strconv.FormatFloat(value.(float64), 'f', -1, 64))
		default:
			return fmt.Errorf("Error unkown type %v for Variable %s", i.varType, i.name)
		}
	}

	for i := range src.stateExpressions {
		value, err := GetValueFromGenericJSONMap(json, src.stateExpressions[i].Field)
		if err != nil {
			return err
		}
		if evaluateStateExpression(value, src.stateExpressions[i]) {
			c <- i
		}
	}
	return nil
}

func evaluateStateExpression(currentValue interface{}, se stateExpression) bool {
	if reflect.TypeOf(currentValue) != reflect.TypeOf(se.Value) {
		return false
	} else {
		switch currentValue.(type) {
		case bool:
			return evaluateBoolean(currentValue.(bool), se.Value.(bool), se.Operator)
		case float64:
			return evaluateFloat(currentValue.(float64), se.Value.(float64), se.Operator)
		case string:
			return evaluateString(currentValue.(string), se.Value.(string), se.Operator)
		}
	}
	return false
}

func evaluateBoolean(x bool, y bool, operator string) bool {
	switch operator {
	case "eq":
		return x == y
	case "ne":
		return x != y
	}
	return false
}

func evaluateFloat(x float64, y float64, operator string) bool {
	switch operator {
	case "eq":
		return x == y
	case "ne":
		return x != y
	case "gt":
		return x > y
	case "ge":
		return x >= y
	case "se":
		fallthrough
	case "le":
		return x <= y
	case "st":
		fallthrough
	case "lt":
		return x < y
	}
	return false
}

func evaluateString(x string, y string, operator string) bool {
	switch operator {
	case "eq":
		return x == y
	case "ne":
		return x != y
	}
	return false
}

func (sr *StateRetriever) init(c chan *State) {
	for i := range sr.states {
		go sr.states[i].runLoop(c)
	}
}

func (src *StateRetrieveCall) runLoop(c chan *State) {
	err := src.retrieve(c)
	if err != nil {
		fmt.Println(err)
	}
	t := time.Tick(time.Duration(src.interval) * time.Second)
	for _ = range t {
		err := src.retrieve(c)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func NewStateRetrieveCall(url string, interval int) *StateRetrieveCall {
	variables := make(map[*Variable]string)
	stateExpressions := make(map[*State]stateExpression)
	return &StateRetrieveCall{variables: variables, url: url, interval: interval, stateExpressions: stateExpressions}
}
