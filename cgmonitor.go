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
	Name     string     //The miners name
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

	fmt.Println("Number of config miners", len(config.Miners))

	//miners := make([]*MinerInformation, len(config.Miners))
	var miners []*MinerInformation

	//Start to grab information from every miner
	for minerName, miner := range config.Miners {
		log.Printf("Server: %s(%s)\n", minerName, miner.IP)
		//Create a new miner struct and add the name
		minerStructTemp := MinerInformation{}
		minerStructTemp.Name = minerName

		//Add save it
		miners = append(miners, &minerStructTemp)

		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10, &minerStructTemp)
	}

	fmt.Println("Number of miners:", len(miners))

	//Loop for ever
	for {
		//Iterate over each miner reponce and print it
		for _, minerInfo := range miners {
			var minerStructTemp = *minerInfo
			//Lock it
			minerStructTemp.Mu.Lock()
			//Read it
			//log.Println(*minerInfo.Name)
			log.Println("Main:", minerStructTemp.Name)
			log.Println("Main:", minerStructTemp.Hashrate)
			//log.Println("")
			//Unlock it
			minerStructTemp.Mu.Unlock()
		}
		//Sleep for some time
		time.Sleep(2 * time.Second)
	}
}
