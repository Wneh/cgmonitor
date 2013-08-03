package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//Struct representing the config.toml
type tomlConfig struct {
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
	Name     string         //The miners name
	Version  string         //Version responce
	SumWrap  SummaryWrapper //Summary
	DevsWrap DevsWrapper    //Devs
	Client   *Client        //RPC Client that holds information
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
var miners map[string]*MinerInformation

func main() {
	log.Println("Starting server...")

	//Create a waitgroup
	wg := new(sync.WaitGroup)

	miners = make(map[string]*MinerInformation)

	//Check that config file exists
	configExists()

	log.Println("Begin reading config file...")
	//Start by reading the config file
	var config tomlConfig
	_, err := toml.DecodeFile("config.toml", &config)
	//Check for errors
	if err != nil {
		log.Println(err)
		return
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
	if _, err := os.Stat("config.toml"); err != nil {
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
	b := []byte("webserverport = 8080\n\n[miners]\n    [miners.alpha]\n    ip = \"127.0.0.1:4028\"]\n    threshold = 0.1")

	err := ioutil.WriteFile("config.toml", b, 0644)
	if err != nil {
		panic(err)
	}
}
