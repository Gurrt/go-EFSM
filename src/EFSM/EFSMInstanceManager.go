package EFSM

import (
	"fmt"
	"sort"
)

type efsmInstance struct {
	efsm               *EFSM
	stateChangeChannel chan *State
	shutdownChannel    chan bool
}

type EFSMInstanceManager struct {
	efsms                  map[string]*efsmInstance
	instanceRetriever      *InstanceRetriever
	latestInstanceSlice    []string
	updatedInstanceChannel chan []string
	shutdownChannel        chan struct{}
	template               root
}

func NewEFSMInstanceManager(ir *InstanceRetriever, template root) *EFSMInstanceManager {
	efsms := make(map[string]*efsmInstance)
	uic := make(chan []string)
	sc := make(chan struct{})
	return &EFSMInstanceManager{efsms: efsms, instanceRetriever: ir, updatedInstanceChannel: uic, shutdownChannel: sc, template: template}
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
	value, found := eim.efsms[id]
	if !found {
		return fmt.Errorf("Error EFSM with ID %s not found", id)
	}
	error := value.efsm.ExecuteFunction(function)
	if error != nil {
		return error
	}
	return nil
}

func (eim *EFSMInstanceManager) Print() {
	fmt.Println("\n", eim.template.Info.Title, " - ", eim.template.Info.Version, "\n BASE URL: ", eim.template.Info.ApiBase, "\n\n EFSM's:")
	var firstKey string
	var keys []string
	for i := range eim.efsms {
		if len(firstKey) == 0 {
			firstKey = i
		}
		efsm := eim.efsms[i].efsm
		fmt.Printf("<%s> Current state: %s\n", i, efsm.currentState.name)

		fmt.Printf("    Variables:\n")
		// Sort keys of map so the order of the variables is always the same
		if len(keys) == 0 {
			for k := range efsm.variableMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		}
		for j := range keys {
			fmt.Printf("     %s : %s\n", keys[j], efsm.variableMap[keys[j]].value)
		}
	}
	fmt.Printf("\nFunctions:\n")
	efsm := eim.efsms[firstKey].efsm
	for i := range efsm.functions {
		fmt.Printf(" %s\n", efsm.functions[i].name)
		fmt.Printf("  Transitions:\n")
		for j := range efsm.functions[i].transitions {
			fmt.Printf("   %s\n", efsm.functions[i].transitions[j].toString())
		}
		fmt.Println("")
	}
}

func (eim *EFSMInstanceManager) handleNewInstanceList(instances []string) {
	foundKeys := make(map[string]bool)
	for key := range eim.efsms {
		foundKeys[key] = false
	}
	for i := range instances {
		_, found := eim.efsms[instances[i]]
		if found {
			foundKeys[instances[i]] = true
		} else {
			eim.AddNewEFSM(instances[i])
		}
	}
	// EFSMS that can't be reached anymore
	for key, value := range foundKeys {
		if !value {
			eim.efsms[key].efsm.Kill()
		}
	}
}

func (eim *EFSMInstanceManager) AddNewEFSM(id string) error {
	if _, found := eim.efsms[id]; found == true {
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
		temp.name = fo.Name
		temp.apiUrl = replaceIdInUrl(id, (eim.template.Info.ApiBase + fo.ApiPath))
		temp.apiContentType = fo.ApiContentType
		temp.apiBody = fo.ApiBody
		temp.apiMethod = fo.ApiMethod

		if fo.Variable != "" {
			temp.variable = efsm.addVariable(fo.Variable)
		}
		for j := range fo.Transitions {
			t := fo.Transitions[j]
			trans, err := efsm.newTransition(t.From, t.To)
			if err != nil {
				return err
			}
			temp.transitions = append(temp.transitions, trans)
		}
		err := efsm.addFunction(temp)
		if err != nil {
			return err
		}
	}

	ch1 := make(chan *State)
	ch2 := make(chan bool)
	newInstance := &efsmInstance{efsm: efsm, stateChangeChannel: ch1, shutdownChannel: ch2}
	eim.efsms[id] = newInstance
	newInstance.efsm.Init()
	return nil
}
