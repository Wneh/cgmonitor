package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path/filepath"
)

//Precache all templates in folder templates at start
var templates = template.Must(template.ParseFiles(filepath.Join("templates", "miners.html"), 
												  filepath.Join("templates", "index.html"),
												  filepath.Join("templates", "miner.html")))

//Starts the webserver
func webServerMain() {

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/miner/{key:[a-zA-Z0-9]+}", MinerHandler)
	r.HandleFunc("/miners", MinersHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web-root/")))
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

//Request handler for a single miner information
func MinerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	//Get the array that hold the information about the devs
	miners[key].DevsWrap.Mu.RLock()
	tempDevs := miners[key].DevsWrap.Devs
	miners[key].DevsWrap.Mu.RUnlock()

	err := templates.ExecuteTemplate(w, "miner.html", tempDevs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//Request handler for a creatin summary for all miners
func MinersHandler(w http.ResponseWriter, r *http.Request) {
	//Generate the correct structure for the template
	tempMiners := createMinersTemplate()

	err := templates.ExecuteTemplate(w, "miners.html", tempMiners)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createMinersTemplate() MinersTemplate {
	var rows []MinerRow

	for _, value := range miners {

		var minerStructTemp = *value

		//Lock it
		minerStructTemp.SumWrap.Mu.RLock()
		//Grab the SummaryObject
		//I think it will always be only 1 object of them
		//so just take index 0 in Summary	
		var summary = &minerStructTemp.SumWrap.Summary.Summary[0]

		//Create a new row and add some infomation
		var row = MinerRow{minerStructTemp.Name, summary.Accepted, summary.Rejected, summary.MHSAv, summary.BestShare}

		rows = append(rows, row)

		//Unlock it
		minerStructTemp.SumWrap.Mu.RUnlock()
	}
	return MinersTemplate{rows}
}

//Default handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type MinersTemplate struct {
	Rows []MinerRow
}

type MinerRow struct {
	Name      string
	Accepted  int
	Rejected  int
	MHSAv     float64
	BestShare int
}
