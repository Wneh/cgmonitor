package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
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

func main() {

	//Start by reading the config file
	var config tomlConfig
	_, err := toml.DecodeFile("config.toml", &config)
	//Check for errors
	if err != nil {
		fmt.Println(err)
		return
	}

	//No errors from loading the config file to start
	//fetching information from each miner
	for minerName, miner := range config.Miners {
		fmt.Printf("Server: %s(%s)\n", minerName, miner.IP)
		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10)
	}
	for {
		fmt.Println("spam")
		time.Sleep(10 * time.Second)
	}
}
