package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"net"
)

type tomlConfig struct {
	Miners map[string]miner
}

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

	for minerName, miner := range config.Miners {
		fmt.Printf("Server: %s(%s)\n", minerName, miner.ip)
	}

	conn, err := net.Dial("tcp", "192.168.1.102:4028")

	c := Client{"Foo", "127.0.0.1", conn, 10}
	fmt.Println("Client: ", c)

	if err != nil {
		panic(err)
	}
	fmt.Fprintf(conn, "{\"command\":\"version\"}")
	status, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Println(status)
}
