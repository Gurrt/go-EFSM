package EFSM

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type root struct {
	Classes []classObject
}

type classObject struct {
	Info      infoObject
	Instances instanceObject
	Profiles  []map[string]string
	Functions []functionObject
	Sync      []apiStateObject
}

type apiStateObject struct {
	ApiPath   string
	Variables map[string]string
	States    map[string]stateExpression
	Interval  int
}

type stateExpression struct {
	Field    string
	Operator string
	Value    interface{}
}

type infoObject struct {
	Title   string
	Version string
	ApiBase string
}

type instanceObject struct {
	ApiPath        string
	APiContentType string
	ApiMethod      string
	IdLocation     string
	Interval       int
}

type idObject struct {
	Type     string
	Location string
}

type functionObject struct {
	Name           string
	Transitions    []transitionObject
	Variable       string
	ApiPath        string
	ApiContentType string
	ApiBody        string
	ApiMethod      string
}

type transitionObject struct {
	From string
	To   string
}

func GetGenericJSONMap(bytes []byte) (map[string]interface{}, error) {
	generic := make(map[string]interface{})
	err := json.Unmarshal(bytes, &generic)
	if err != nil {
		fmt.Println("Incoming JSON: ")
		return nil, err
	}
	return generic, nil
}

// TODO: Splitting and then joining again is not very efficient, could replace with
// inner recursive function that works directly on the array of paths
func GetValueFromGenericJSONMap(genJson map[string]interface{}, path string) (interface{}, error) {
	paths := strings.Split(path, ".")
	if genJson[paths[0]] == nil {
		return nil, fmt.Errorf("Component %s from path %s not found in JSON %v\n", paths[0], path, genJson)
	} else if len(paths) == 1 {
		return genJson[paths[0]], nil
	} else {
		return GetValueFromGenericJSONMap(genJson[paths[0]].(map[string]interface{}), strings.Join(paths[1:], "."))
	}
}

func GetMultipleValuesFromGenericJSONMap(genJson map[string]interface{}, path string) ([]string, error) {
	paths := strings.Split(path, ".")
	var results []string
	// Special mode to iterate over key value
	if paths[0] == "$key" {
		if len(paths) == 1 {
			for k := range genJson {
				results = append(results, k)
			}
			return results, nil
		} else {
			for _, v := range genJson {
				new, err := GetMultipleValuesFromGenericJSONMap(v.(map[string]interface{}), strings.Join(paths[1:], "."))
				if err != nil {
					fmt.Print(err)
				}
				results = append(results, new...)
			}
		}
	} else {
		if genJson[paths[0]] == nil {
			return nil, fmt.Errorf("Component %s from path %s not found in JSON %v\n", paths[0], path, genJson)
		} else if len(paths) == 1 {
			result := genJson[paths[0]]
			switch result.(type) {
			case string:
				return []string{result.(string)}, nil
			case bool:
				return []string{strconv.FormatBool(result.(bool))}, nil
			case float64:
				return []string{strconv.FormatFloat(result.(float64), 'f', -1, 64)}, nil
			default:
				return nil, fmt.Errorf("Error unkown type for variable %v", path)
			}
		} else {
			return GetMultipleValuesFromGenericJSONMap(genJson[paths[0]].(map[string]interface{}), strings.Join(paths[1:], "."))
		}
	}
	return nil, nil
}

func replaceIdInUrl(id string, url string) string {
	return strings.Replace(url, "$id", id, -1)
}

func FromJSONFile(filename string) ([]*EFSMInstanceManager, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var r root
	if err := json.Unmarshal(contents, &r); err != nil {
		fmt.Print("Error unmarshalling!\n")
		return nil, err
	}

	var instanceManagers []*EFSMInstanceManager

	for i := range r.Classes {
		class := r.Classes[i]

		profiles := make(map[string]*Profile)
		for j := range class.Profiles {
			profile := class.Profiles[j]
			var id string
			conversions := make(map[string]string)
			for key := range profile {
				if key == "id" {
					id = profile[key]
				} else {
					conversions[key] = profile[key]
				}
			}
			pro := &Profile{Id: id, Conversions: conversions}
			profiles[id] = pro
		}

		ir := &InstanceRetriever{url: class.Info.ApiBase + class.Instances.ApiPath,
			interval:  class.Instances.Interval,
			location:  class.Instances.IdLocation,
			apiMethod: class.Instances.ApiMethod,
			apiBody:   ""}

		eim := NewEFSMInstanceManager(ir, profiles, class)
		eim.Init()
		instanceManagers = append(instanceManagers, eim)
	}

	return instanceManagers, nil
}
