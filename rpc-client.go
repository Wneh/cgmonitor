package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Client struct {
	Name            string            //Name of the miner
	IP              string            //Ip to the cgminer including port
	Conn            net.Conn          //Connection made with net.Dial
	RefreshInterval int               //Seconds between fetching information
	MinerInfo       *MinerInformation //Struc to put the answers for the webserver	
}

//Main function for fetching information from one client
func rpcClient(name, ip string, refInt int, minerInfo *MinerInformation) {
	//Add everything except the connection
	c := Client{name, ip, nil, refInt, minerInfo}

	fmt.Println(c)

	//Create the connection
	conn, err := net.Dial("tcp", c.IP)

	//Check for errors
	if err != nil {
		log.Println(err)
		return
	}

	c.Conn = conn
	//c.Conn = createConnection(c.IP)
	//Send test command
	//fmt.Println(sendCommand(&c.Conn, "{\"command\":\"version\"}"))

	//Continue asking the miner for the hashrate
	for {
		//Send and print the reponce
		fmt.Println(sendCommand(&c.Conn, "{\"command\":\"summary\"}"))
		/* 
		 * Note:
		 *
		 * It seems that cgminer close the tcp connection
		 * after each call so we need to reset it for
		 * the next rpc-call
		 */
		conn.Close()
		//Create the new connection
		c.Conn = createConnection(c.IP)

		//Sleep for the a while
		time.Sleep(10 * time.Second)
	}
}

func createConnection(ip string) net.Conn {
	conn, err := net.Dial("tcp", ip)

	//Check for errors
	if err != nil {
		log.Println(err)
		return nil
	}

	return conn
}

func sendCommand(conn *net.Conn, cmd string) string {
	//Write the command to the socket
	fmt.Fprintf(*conn, cmd)
	//Read the response
	status, err := bufio.NewReader(*conn).ReadString('\n')
	//Check for any errors
	if err != nil {
		//Check for errors
		if err == io.EOF {
			//cgminer seems to close after each call so close the socket and dial it again
			//log.Println("Sending command error: ", err)
		} else {
			//If the error is not EOF then warn about it
			log.Println("Sending command error: ", err)
		}
	}
	//Return the status we got from the server
	return status
}
