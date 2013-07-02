package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"sync"
)

//Struct representing the config.toml
type tomlConfig struct {
	Miners map[string]miner //Key is the miner name.
}

//Struct for config file type [miners.<foo>] 
type miner struct {
	IP string
}

type MinerInformation struct {
	Name    string          //The miners name
	Version string          //Version responce
	SumWrap SummaryWrapper //Summary
	DevsWrap 	DevsWrapper	//Devs
}

type SummaryWrapper struct {
	Mu sync.RWMutex
	Summary SummaryResponse
}

type DevsWrapper struct {
	Mu sync.RWMutex
	Devs DevsResponse
}

//Global variabels
//var miners []*MinerInformation
var miners map[string]*MinerInformation

func main() {
	log.Println("Starting server...")

	miners = make(map[string]*MinerInformation)

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
		log.Printf("Server: %s(%s)\n", minerName, miner.IP)
		//Create a new miner struct and add the name
		minerStructTemp := MinerInformation{}
		minerStructTemp.Name = minerName

		//Add save it
		miners[minerName] = &minerStructTemp

		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10, &minerStructTemp)
		log.Printf("    Started %s(%s) thread", minerName, miner.IP)
	}
	log.Println("...done starting rpc-client threads")

	log.Println("Starting web server")
	//time.Sleep(5 * time.Second)
	webServerMain()
}
