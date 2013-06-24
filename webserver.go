package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
)

//Starts a webserver
func webServerMain() {

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/miner/{key}", MinerHandler)
	r.HandleFunc("/miners", MinersHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web-root/")))
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

//Request handler for a single miner information
func MinerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	fmt.Println(key)

	fmt.Fprintf(w, "The miner your looking for is: %s", key)
}

//Request handler for a creatin summary for all miners
func MinersHandler(w http.ResponseWriter, r *http.Request) {
	//Generate the correct structure for the template
	tempMiners := createMinersTemplate()

	// fmt.Println("Miners: ",&tempMiners.Rows)

	for _,value := range tempMiners.Rows {
		fmt.Printf("%s\n",value)
	}

	t, err := template.ParseFiles("miners.html")
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, tempMiners)

	//fmt.Fprintf(w, "%s", result)
}

func createMinersTemplate() (MinersTemplate){
	var rows []MinerRow

	for _, value := range miners {
		
		var minerStructTemp = *value

		//Lock it
		minerStructTemp.Mu.Lock()
		//Grab the SummaryObject
		//I think it will always be only 1 object of them
		//so just take index 0 in Summary	
		var summary = &minerStructTemp.Summary.Summary[0]

		//Create a new row and add some infomation
		var row = MinerRow{minerStructTemp.Name,summary.Accepted,summary.Rejected,summary.MHSAv,summary.BestShare}

		rows = append(rows, row)

		//Unlock it
		minerStructTemp.Mu.Unlock()
	}

	return MinersTemplate{rows}
}

//Default handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//respons := ""
	fmt.Fprintf(w, "Start page!!")
}

type MinersTemplate struct{
	Rows []MinerRow
}

type MinerRow struct{
	Name string
	Accepted int
	Rejected int
	MHSAv float64
	BestShare int
}
