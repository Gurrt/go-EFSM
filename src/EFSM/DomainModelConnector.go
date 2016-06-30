package EFSM

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DomainModelConnector struct {
	url string
}

func (dmc *DomainModelConnector) getVariablesForProfiles(profiles map[string]*Profile) map[string]Variable {
	variables := make(map[string]Variable)
	for key := range profiles {
		json, err := dmc.getJsonObjectForProfile(key)
		if err != nil {
			fmt.Print(err)
			return nil
		}
		fields := json["Fields"].(map[string]interface{})
		for variable := range fields {
			varObj := fields[variable].(map[string]interface{})
			varId := key + "." + variable
			variables[varId] = Variable{Name: variable, Profile: key, VarType: varObj["Type"].(string), Value: ""}
		}
	}
	return variables
}

func (dmc *DomainModelConnector) getJsonObjectForProfile(profile string) (map[string]interface{}, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest("GET", dmc.url+"profiles/"+profile, nil)

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

	json, err := GetGenericJSONMap(body)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (dmc *DomainModelConnector) convertToDomainModel(profile string, variable string, value string, encoding string) (string, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest("GET", dmc.url+"profiles/"+profile+"/"+variable+"/convert?from="+encoding+"&value="+value, nil)

	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	json, err := GetGenericJSONMap(body)
	if err != nil {
		return "", err
	}

	result := json["result"]

	switch result.(type) {
	case string:
		return result.(string), nil
	case bool:
		return strconv.FormatBool(result.(bool)), nil
	case float64:
		return strconv.FormatFloat(result.(float64), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("Unkown JSON Type")
	}
}

func (dmc *DomainModelConnector) convertFromDomainModel(profile string, variable string, value string, encoding string) (string, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest("GET", dmc.url+"profiles/"+profile+"/"+variable+"/convert?to="+encoding+"&value="+value, nil)

	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	json, err := GetGenericJSONMap(body)
	if err != nil {
		return "", err
	}

	result := json["result"]

	switch result.(type) {
	case string:
		return result.(string), nil
	case bool:
		return strconv.FormatBool(result.(bool)), nil
	case float64:
		return strconv.FormatFloat(result.(float64), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("Unkown JSON Type")
	}
}
