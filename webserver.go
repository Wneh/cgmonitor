package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
)

//Precache all templates in folder templates at start
var templates = template.Must(template.ParseFiles(filepath.Join("templates", "miners.html"),
	filepath.Join("templates", "index.html"),
	filepath.Join("templates", "miner.html")))

//Starts the webserver
func webServerMain(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/miner/{key:[a-zA-Z0-9]+}", MinerHandler)
	r.HandleFunc("/miners", MinersHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web-root/")))
	http.Handle("/", r)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
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
		//Add it
		rows = append(rows, minerStructTemp.SumWrap.SummaryRow)
		//Unlock it
		minerStructTemp.SumWrap.Mu.RUnlock()
	}
	return MinersTemplate{rows}
}

//Default handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	hw := HomeWrapper{}

	//Calculate the total hashrate
	for _, value := range miners {
		var minerStructTemp = *value

		hw.TotalMHS += minerStructTemp.SumWrap.SummaryRow.MHSAv
	}

	err := templates.ExecuteTemplate(w, "index.html", hw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type HomeWrapper struct {
	TotalMHS float64
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
