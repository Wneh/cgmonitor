package main

import (
	"github.com/BurntSushi/toml"
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
	IP string
}

type MinerInformation struct {
	Name     string         //The miners name
	Version  string         //Version responce
	SumWrap  SummaryWrapper //Summary
	DevsWrap DevsWrapper    //Devs
}

type SummaryWrapper struct {
	Mu         sync.RWMutex
	Summary    SummaryResponse
	SummaryRow MinerRow
}

type DevsWrapper struct {
	Mu   sync.RWMutex
	Devs DevsResponse
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
		go rpcClient(minerName, miner.IP, 10, &minerStructTemp, wg)
	}
	log.Println("...Waiting for every thread to be started")
	wg.Wait()

	log.Println("All thread started, starting web server")
	webServerMain(config.Webserverport)
}

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

func createExampleConf() {

}
