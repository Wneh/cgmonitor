package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func webServerMain() {
	// http.HandleFunc("/", handler)
	// http.HandleFunc("/", http.FileServer(http.Dir(./)))
	// http.ListenAndServe(":8080", nil)

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/products/{key}", ProductsHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	fmt.Println(key)

	fmt.Fprintf(w, "The product: %s", key)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//respons := ""

	fmt.Fprintf(w, "Start page!!")
	//fmt.Fprintf(w, "")
	// for _, minerInfo := range miners {
	// 	var minerStructTemp = *minerInfo
	// 	//Lock it
	// 	minerStructTemp.Mu.Lock()
	// 	//Read it
	// 	//log.Println(*minerInfo.Name)
	// 	//fmt.Fprintf(w, "Main:", miners)
	// 	//fmt.Fprintf(w, "Main:", minerStructTemp.Hashrate)
	// 	//fmt.Fprintf(w, "")
	// 	respons += minerStructTemp.Name + "\n"
	// 	respons += minerStructTemp.Hashrate + "\n"
	// 	respons += ".html"
	// 	//log.Println("")
	// 	//Unlock it
	// 	minerStructTemp.Mu.Unlock()
	// }
	// fmt.Fprintf(w, respons)
}

// var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

// //Based on the makeHandler from the golang wiki article
// func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		//title := r.URL.Path[lenPath:]
// 		//if !titleValidator.MatchString(title) {
// 		//	http.NotFound(w, r)
// 		//	return
// 		//}
// 		fn(w, r, title)
// 	}
// }
