package EFSM

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type InstanceRetriever struct {
	url       string
	interval  int
	idType    string
	location  string
	apiMethod string
	apiBody   string
}

func (inst *InstanceRetriever) retrieve(c chan []string) {
	client := &http.Client{}

	var req *http.Request
	var err error
	if len(inst.apiBody) > 0 {
		req, err = http.NewRequest(inst.apiMethod, inst.url, strings.NewReader(inst.apiBody))
	} else {
		req, err = http.NewRequest(inst.apiMethod, inst.url, nil)
	}

	if err != nil {
		fmt.Print(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Print(err)
	}

	json, err := GetGenericJSONMap(body)
	if err != nil {
		fmt.Print(err)
	}
	var instances []string
	instances, err = GetMultipleValuesFromGenericJSONMap(json, inst.location)
	if err != nil {
		fmt.Print(err)
	}
	c <- instances
}

func (inst *InstanceRetriever) init(c chan []string, quit chan struct{}) {
	go func() {
		inst.retrieve(c)
		t := time.NewTicker(time.Duration(inst.interval) * time.Second)
		for {
			select {
			case <-quit:
				t.Stop()
				return
			case <-t.C:
				inst.retrieve(c)
			}
		}
	}()
}
