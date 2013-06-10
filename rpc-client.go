package main

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	Name            string   //Name of the miner
	IP              string   //Ip to the cgminer including port
	Conn            net.Conn //Connection made with net.Dial
	RefreshInterval int      //Seconds between fetching information
}

//Main function for fetching information from one client
func rpcClient(name, ip string, refInt int) {
	//Add everything except the connection
	c := Client{name, ip, nil, refInt}

	fmt.Println(c)

	//Create the connection
	conn, err := net.Dial("tcp", c.IP)

	//Check for errors
	if err != nil {
		panic(err)
	}

	//Save it because it went well
	c.Conn = conn
	//Send test command
	fmt.Fprintf(conn, "{\"command\":\"version\"}")
	//Read the response
	status, err := bufio.NewReader(conn).ReadString('\n')
	//Print it
	fmt.Println(status)
}
