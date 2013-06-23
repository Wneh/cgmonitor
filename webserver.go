package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
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

	result := ""

	for _, value := range miners {
		var minerStructTemp = *value

		minerStructTemp.Mu.Lock()
		//Read it
		//log.Println(*minerInfo.Name)
		//log.Println("Main:", minerStructTemp.Name)
		//log.Println("Main:", minerStructTemp.Hashrate)
		fmt.Printf("%v\n", minerStructTemp.Summary.Summary[0].MHSAv)

		result += minerStructTemp.Name + ": " + fmt.Sprintf("%g", minerStructTemp.Summary.Summary[0].MHSAv) + "\n"

		//log.Println("")
		//Unlock it
		minerStructTemp.Mu.Unlock()

	}

	fmt.Fprintf(w, "%s", result)
}

//Default handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//respons := ""
	fmt.Fprintf(w, "Start page!!")
}
