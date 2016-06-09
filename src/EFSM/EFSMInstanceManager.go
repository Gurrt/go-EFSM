package EFSM

import (
	"fmt"
	"sort"
)

type efsmInstance struct {
	Efsm               *EFSM
	stateChangeChannel chan *State
	shutdownChannel    chan bool
}

type EFSMInstanceManager struct {
	Efsms                  map[string]*efsmInstance
	instanceRetriever      *InstanceRetriever
	latestInstanceSlice    []string
	updatedInstanceChannel chan []string
	shutdownChannel        chan struct{}
	template               classObject
}

type DetailedClassJSON struct {
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	Instances []InstanceJSON `json:"instances"`
	States    []string       `json:"states"`
	Functions []FunctionJSON `json:"functions"`
}

func (eim *EFSMInstanceManager) Serialize() DetailedClassJSON {
	var instances []InstanceJSON
	var firstKey string
	for k, v := range eim.Efsms {
		if firstKey == "" {
			firstKey = k
		}
		instances = append(instances, v.Efsm.Serialize())
	}

	// TODO: Move states and functions to the EFSM Instance Manager
	// Grab the states and functions from one of the EFSM's
	var states []string

	for _, state := range eim.Efsms[firstKey].Efsm.States {
		states = append(states, state.Name)
	}

	var functions []FunctionJSON
	funcArray := eim.Efsms[firstKey].Efsm.Functions
	for i := range funcArray {
		functions = append(functions, funcArray[i].Serialize())
	}

	return DetailedClassJSON{Name: eim.template.Info.Title,
		Version:   eim.template.Info.Version,
		States:    states,
		Functions: functions,
		Instances: instances}
}

func NewEFSMInstanceManager(ir *InstanceRetriever, template classObject) *EFSMInstanceManager {
	efsms := make(map[string]*efsmInstance)
	uic := make(chan []string)
	sc := make(chan struct{})
	return &EFSMInstanceManager{Efsms: efsms, instanceRetriever: ir, updatedInstanceChannel: uic, shutdownChannel: sc, template: template}
}

func (eim *EFSMInstanceManager) Init() {
	eim.instanceRetriever.init(eim.updatedInstanceChannel, eim.shutdownChannel)
	firstUpdated := make(chan struct{})
	go func() {
		first := true
		for {
			select {
			case x := <-eim.updatedInstanceChannel:
				eim.handleNewInstanceList(x)
				if first {
					close(firstUpdated)
					first = false
				}
			}
		}
	}()
	_ = <-firstUpdated
}

func (eim *EFSMInstanceManager) Exec(id string, function string) error {
	value, found := eim.Efsms[id]
	if !found {
		return fmt.Errorf("Error EFSM with ID %s not found", id)
	}
	error := value.Efsm.ExecuteFunction(function)
	if error != nil {
		return error
	}
	return nil
}

func (eim *EFSMInstanceManager) Print(index int) {
	fmt.Println("\n", "Index [", index, "]", eim.template.Info.Title, " - ", eim.template.Info.Version, "\n BASE URL: ", eim.template.Info.ApiBase, "\n\n EFSM's:")
	var firstKey string
	var keys []string
	for i := range eim.Efsms {
		if len(firstKey) == 0 {
			firstKey = i
		}
		efsm := eim.Efsms[i].Efsm
		fmt.Printf("<%s> Current state: %s\n", i, efsm.CurrentState.Name)

		fmt.Printf("    Variables:\n")
		// Sort keys of map so the order of the variables is always the same
		if len(keys) == 0 {
			for k := range efsm.VariableMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		}
		for j := range keys {
			fmt.Printf("     %s : %s\n", keys[j], efsm.VariableMap[keys[j]].Value)
		}
	}
	fmt.Printf("\nFunctions:\n")
	efsm := eim.Efsms[firstKey].Efsm
	for i := range efsm.Functions {
		fmt.Printf(" %s\n", efsm.Functions[i].Name)
		fmt.Printf("  Transitions:\n")
		for j := range efsm.Functions[i].Transitions {
			fmt.Printf("   %s\n", efsm.Functions[i].Transitions[j].toString())
		}
		fmt.Println("")
	}
}

func (eim *EFSMInstanceManager) handleNewInstanceList(instances []string) {
	foundKeys := make(map[string]bool)
	for key := range eim.Efsms {
		foundKeys[key] = false
	}
	for i := range instances {
		_, found := eim.Efsms[instances[i]]
		if found {
			foundKeys[instances[i]] = true
		} else {
			eim.AddNewEFSM(instances[i])
		}
	}
	// EFSMS that can't be reached anymore
	for key, value := range foundKeys {
		if !value {
			eim.Efsms[key].Efsm.Kill()
		}
	}
}

func (eim *EFSMInstanceManager) AddNewEFSM(id string) error {
	if _, found := eim.Efsms[id]; found == true {
		return fmt.Errorf("EFSM with id %s already exists, not adding a new EFSM with the same ID", id)
	}
	efsm := NewEFSM(id)
	sr := &StateRetriever{}

	stateCalls := eim.template.Sync
	for i := range stateCalls {
		src := NewStateRetrieveCall(replaceIdInUrl(id, (eim.template.Info.ApiBase+stateCalls[i].ApiPath)), stateCalls[i].Interval)
		for key, value := range stateCalls[i].Variables {
			src.variables[efsm.addVariable(key)] = value
		}

		for key, value := range stateCalls[i].States {
			state, err := efsm.addState(key)
			if err != nil {
				return err
			}
			src.stateExpressions[state] = value
		}

		sr.states = append(sr.states, src)
	}
	efsm.stateRetriever = sr

	for i := range eim.template.Functions {
		fo := &eim.template.Functions[i]
		temp := new(Function)
		temp.Name = fo.Name
		temp.apiUrl = replaceIdInUrl(id, (eim.template.Info.ApiBase + fo.ApiPath))
		temp.apiContentType = fo.ApiContentType
		temp.apiBody = fo.ApiBody
		temp.apiMethod = fo.ApiMethod

		if fo.Variable != "" {
			temp.Variable = efsm.addVariable(fo.Variable)
		}
		for j := range fo.Transitions {
			t := fo.Transitions[j]
			trans, err := efsm.newTransition(t.From, t.To)
			if err != nil {
				return err
			}
			temp.Transitions = append(temp.Transitions, trans)
		}
		err := efsm.addFunction(temp)
		if err != nil {
			return err
		}
	}

	ch1 := make(chan *State)
	ch2 := make(chan bool)
	newInstance := &efsmInstance{Efsm: efsm, stateChangeChannel: ch1, shutdownChannel: ch2}
	eim.Efsms[id] = newInstance
	newInstance.Efsm.Init()
	return nil
}
