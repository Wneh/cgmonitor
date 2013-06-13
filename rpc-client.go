package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
		log.Println(err)
		return
	}

	//Save it because it went well
	c.Conn = conn
	//Send test command
	fmt.Println(sendCommand(&c.Conn, "{\"command\":\"version\"}"))
}

func sendCommand(conn *net.Conn, cmd string) string {
	//Write the command to the socket
	fmt.Fprintf(*conn, cmd)
	//Read the response
	status, err := bufio.NewReader(*conn).ReadString('\n')
	//Check for any errors
	if err != nil {
		//If the error is not EOF then warn about it
		if err != io.EOF {
			log.Println("Sending command error: ", err)
		}
	}
	//Return the status we got from the server
	return status
}
