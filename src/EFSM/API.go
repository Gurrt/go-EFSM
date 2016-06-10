package EFSM

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type RuntimeInfo struct {
	classes []*EFSMInstanceManager
}

type TopLevelJSON struct {
	Classes []ClassJSON `json:"classes"`
}

type ClassJSON struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

type ErrorJSON struct {
	Error string `json:"error"`
}

var serverPort int = 8080

func (runtime *RuntimeInfo) generateTopLevelJSON(r *http.Request) TopLevelJSON {
	var classes []ClassJSON
	for i := range runtime.classes {
		class := ClassJSON{ID: i,
			Name:    runtime.classes[i].template.Info.Title,
			Version: runtime.classes[i].template.Info.Version,
			URL:     fmt.Sprintf("http://%s/classes/%d", r.Host, i)}
		classes = append(classes, class)
	}
	return TopLevelJSON{Classes: classes}
}

func (runtime *RuntimeInfo) generateClassJSON(index int, baseURL string) DetailedClassJSON {
	return runtime.classes[index].Serialize(baseURL)
}

func StartAPI(classes []*EFSMInstanceManager) {
	runtime := &RuntimeInfo{classes: classes}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/classes/", runtime.getClasses).Methods("GET")
	router.HandleFunc("/classes/{id}", runtime.getClass).Methods("GET")
	router.HandleFunc("/classes/{id}/{func}", runtime.executeFunction).Methods("GET")

	fmt.Println("Starting API Server @ port 8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), handlers.CORS()(router)))
}

func (runtime *RuntimeInfo) executeFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	intID, _ := strconv.Atoi(id)
	if intID > len(runtime.classes)-1 || intID < 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(&ErrorJSON{Error: fmt.Sprintf("Class %d does not exist", intID)})
		return
	}

	eim := runtime.classes[intID]
	function := vars["func"]

	query := r.URL.Query()
	ids := query["ids"][0]
	if len(ids) == 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(&ErrorJSON{Error: "Query field 'ids' is empty"})
		return
	}
	idList := strings.Split(ids, ",")
	var value string
	valueArr := query["value"]
	if valueArr != nil {
		value = string(valueArr[0])
	}
	fmt.Println("val", value)
	for i := range idList {
		efsm, found := eim.Efsms[idList[i]]
		if !found {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(&ErrorJSON{Error: fmt.Sprintf("Instance id %s is unknown", idList[i])})
			return
		}
		efsm.Efsm.ExecuteFunction(function, value)
	}

}

func (runtime *RuntimeInfo) getClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	intID, _ := strconv.Atoi(id)
	if intID > len(runtime.classes)-1 || intID < 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(&ErrorJSON{Error: fmt.Sprintf("Class %d does not exist", intID)})
		return
	}
	baseURL := fmt.Sprintf("http://%s/classes/%s", r.Host, id)
	json.NewEncoder(w).Encode(runtime.generateClassJSON(intID, baseURL))
}

func (runtime *RuntimeInfo) getClasses(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(runtime.generateTopLevelJSON(r))
}
