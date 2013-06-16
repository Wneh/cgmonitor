package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
	"sync"
	"time"
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
	Mu       sync.Mutex //So we dont read and write to it at the same time.
	Version  string     //Version responce
	Hashrate string     //Hashrate response for the miner
}

func main() {

	//Start by reading the config file
	var config tomlConfig
	_, err := toml.DecodeFile("config.toml", &config)
	//Check for errors
	if err != nil {
		fmt.Println(err)
		return
	}

	minerStruct := MinerInformation{}

	//No errors from loading the config file to start
	//fetching information from each miner
	for minerName, miner := range config.Miners {
		fmt.Printf("Server: %s(%s)\n", minerName, miner.IP)
		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10, &minerStruct)
	}
	for {
		//Lock it
		minerStruct.Mu.Lock()
		//Read it
		log.Println(minerStruct.Hashrate)
		//Unlock it
		minerStruct.Mu.Unlock()
		//Sleep for some time
		time.Sleep(2 * time.Second)
	}
}
