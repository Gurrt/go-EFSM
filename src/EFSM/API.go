package EFSM

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (runtime *RuntimeInfo) generateClassJSON(index int) DetailedClassJSON {
	return runtime.classes[index].Serialize()
}

func StartAPI(classes []*EFSMInstanceManager) {
	runtime := &RuntimeInfo{classes: classes}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/classes/", runtime.getClasses).Methods("GET")
	router.HandleFunc("/classes/{id}", runtime.getClass).Methods("GET")

	fmt.Println("Starting API Server @ port 8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), router))
}

func (runtime *RuntimeInfo) getClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	strId, _ := strconv.Atoi(id)
	json.NewEncoder(w).Encode(runtime.generateClassJSON(strId))
}

func (runtime *RuntimeInfo) getClasses(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(runtime.generateTopLevelJSON(r))
}
