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

//Default handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//respons := ""
	fmt.Fprintf(w, "Start page!!")
}
