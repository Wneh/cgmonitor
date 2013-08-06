package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//Struct representing the config.toml
type Config struct {
	Webserverport int              //Webserver port
	Miners        map[string]miner //Key is the miner name.
}

//Struct for config file type [miners.<foo>] 
type miner struct {
	IP        string  //The ip
	Threshold float64 //The threshold that the miner cant go below(in mega hashes)
	KeepAlive bool    //If restart command should be send on sick/dead detection
}

//"Top" struct that represent each miner.
//This struct is used in the var miners that is called from here and there
type MinerInformation struct {
	Name         string         //The miners name
	SumWrap      SummaryWrapper //Summary
	DevsWrap     DevsWrapper    //Devs
	Client       *Client        //RPC Client that holds information
	ClientConfig miner
}

//Wrapper that contain all the information that is used for the summary request
type SummaryWrapper struct {
	Mu         sync.RWMutex    //Mutex lock
	Summary    SummaryResponse //The response parsed to a struct
	SummaryRow MinerRow        //The response converted to rows for the web page
}

//Wrapper that contain all the information that is used for the devs request
type DevsWrapper struct {
	Mu   sync.RWMutex //The mutex lock
	Devs DevsResponse //The response from devs in json format 
}

//Global variabels
//Map containing with the miner name as key
var miners map[string]*MinerInformation

//Logger that both write to console and file

func main() {
	//Open the log file
	logf, err := os.OpenFile("cgmonitor.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
	if err != nil {
		log.Fatal(err)
	}
	//Make sure that we close it later on
	defer logf.Close()

	//Set log to both file and console
	log.SetOutput(io.MultiWriter(logf, os.Stdout))

	log.Println("Starting server...")

	//Create a waitgroup
	wg := new(sync.WaitGroup)

	miners = make(map[string]*MinerInformation)

	//Check that config file exists
	configExists()

	log.Println("Begin reading config file...")
	//Start by reading the config file
	var config Config

	//Read the config file
	b, err := ioutil.ReadFile("cgmonitor.conf")
	if err != nil {
		panic(err)
	}
	//Parse the raw json to struct
	err = json.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}
	log.Println("...done reading config file")

	log.Println("Begin starting rpc-client threads...")
	//Start to grab information from every miner
	for minerName, miner := range config.Miners {
		//Add two(one for each thread) to the WaitGroup for each miner
		wg.Add(2)

		log.Printf("Starting: %s(%s)\n", minerName, miner.IP)
		//Create a new miner struct and add the name
		minerStructTemp := MinerInformation{}
		minerStructTemp.Name = minerName
		minerStructTemp.ClientConfig = miner

		//Add save it
		miners[minerName] = &minerStructTemp

		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10, &minerStructTemp, wg, miner.Threshold)
	}
	log.Println("...Waiting for every thread to be started")
	wg.Wait()

	log.Println("All thread started, starting web server")
	webServerMain(config.Webserverport)
}

//Check of the there is a config.toml file is the same folder as the program is runned from
//If not it will create one
func configExists() {
	if _, err := os.Stat("cgmonitor.conf"); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			log.Println("No config file found, creating example config file.")
			createExampleConf()
		} else {
			// other error
		}
	}
}

//Creates a basic config file
func createExampleConf() {
	//Create the Config struct and add some values
	tempConf := Config{8080, make(map[string]miner)}
	tempConf.Miners["alpha"] = miner{"127.0.0.1:4028", 0.1, true}

	//Convert it to json
	b, err := json.MarshalIndent(tempConf, "", "    ")
	if err != nil {
		log.Println("error:", err)
	}

	//And save it to file
	err = ioutil.WriteFile("cgmonitor.conf", b, 0644)
	if err != nil {
		panic(err)
	}
}
