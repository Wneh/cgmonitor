package main

import (
	"fmt"
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
	r.HandleFunc("/miner/{key:[a-zA-Z0-9]+}/onoff", EnableDisableHandler)
	r.HandleFunc("/miner/{key:[a-zA-Z0-9]+}/gpu", GPUHandler)
	r.HandleFunc("/miners", MinersHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web-root/")))
	http.Handle("/", r)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

//Request handler for a single miner information
func MinerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	//Check if a miner with that name exists
	if _, exists := miners[key]; !exists {
		http.Error(w, "404 - No miner with that name", http.StatusNotFound)
		return
	}

	miner := MinerWrapper{}
	miner.Name = key

	//Get the array that hold the information about the devs
	miners[key].DevsWrap.Mu.RLock()
	miner.Devs = miners[key].DevsWrap.Devs
	miners[key].DevsWrap.Mu.RUnlock()
	//fmt.Printf("Onoff: %s\n", miner.Devs.Devs[0].OnOff)

	err := templates.ExecuteTemplate(w, "miner.html", miner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func EnableDisableHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	//Parse the values
	statusNumber, err := strconv.Atoi(r.FormValue("status"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	deviceNumber, err := strconv.Atoi(r.FormValue("device"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enableDisable(statusNumber, deviceNumber, key)

	//And before we redirect we update the devs information
	//But dont to the threshold check
	UpdateDevs(key, false)

	http.Redirect(w, r, "/miner/"+key, http.StatusFound)
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

func GPUHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	//Parse the values
	deviceNumber, err := strconv.Atoi(r.FormValue("device"))
	if err != nil {
		http.Error(w, err.Error()+"1", http.StatusInternalServerError)
		return
	}

	gpuClock, err := strconv.Atoi(r.FormValue("GPUClock"))
	if err != nil {
		http.Error(w, err.Error()+"2", http.StatusInternalServerError)
		return
	}
	//Check so it's in the right interval
	if gpuClock < 150 || gpuClock > 1200 {
		http.Error(w, "GPUClock must be between 150-1200", http.StatusInternalServerError)
	}

	gpuMemory, err := strconv.Atoi(r.FormValue("MemoryClock"))
	if err != nil {
		http.Error(w, err.Error()+"3", http.StatusInternalServerError)
		return
	}
	//Check so it's in the right interval
	if gpuMemory < 75 || gpuMemory > 1575 {
		http.Error(w, "GPUClock must be between 150-1200", http.StatusInternalServerError)
	}

	/*
	 * Note:
	 * Possible bugg there. It might lose precision 
	 * with the convert! Looking into this later
	 */
	vddc, err := strconv.ParseFloat(r.FormValue("Voltage"), 128)
	if err != nil {
		http.Error(w, err.Error()+"4", http.StatusInternalServerError)
		return
	}
	//Check so it's in the right interval
	if vddc < 0.800 || vddc > 1.250 {
		http.Error(w, "GPUClock must be between 150-1200", http.StatusInternalServerError)
	}

	intensity, err := strconv.Atoi(r.FormValue("Intensity"))
	if err != nil {
		http.Error(w, err.Error()+"5", http.StatusInternalServerError)
		return
	}
	//Check so it's in the right interval
	if intensity < -10 || intensity > 20 {
		http.Error(w, "GPUClock must be between 150-1200", http.StatusInternalServerError)
	}

	config := r.FormValue("Config")

	fmt.Println(gpuClock, deviceNumber, gpuMemory, vddc, intensity)

	setGPUEngine(gpuClock, deviceNumber, key)
	setGPUMemory(gpuMemory, deviceNumber, key)
	setVDDC(vddc, deviceNumber, key)
	setIntensity(intensity, deviceNumber, key)

	//Check if we shold write config file
	if config == "on" {
		writeConfig(key)
	}

	//And before we redirect we update the devs information
	//But dont to the threshold check
	UpdateDevs(key, false)

	http.Redirect(w, r, "/miner/"+key, http.StatusFound)
}

type MinerWrapper struct {
	Name string
	Devs DevsResponse
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
