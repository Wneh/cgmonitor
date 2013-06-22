package main

import (
	"bufio"
	"encoding/json"
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
	//Save it to the struct
	c.Conn = conn

	//Continue asking the miner for the hashrate
	for {
		//Get the new information
		b := []byte(sendCommand(&c.Conn, "{\"command\":\"summary\"}"))
		//fmt.Printf("%#v", temp)

		/*
		 * Check for \x00 to remove
		 */
		if b[len(b)-1] == '\x00' {
			b = b[0 : len(b)-1]
		}

		s := SummaryResponse{}
		err := json.Unmarshal(b, &s)
		//Check for errors
		if err != nil {
			//panic(err)
			fmt.Println(err.Error())
		}

		//Lock because we going to write to the minerInfo
		minerInfo.Mu.Lock()

		//Save the summary
		minerInfo.Summary = s

		//Now unlock
		minerInfo.Mu.Unlock()

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
		time.Sleep(time.Duration(c.RefreshInterval) * time.Second)
	}
}

// Returns a TCP connection to the ip 
func createConnection(ip string) net.Conn {
	conn, err := net.Dial("tcp", ip)

	//Check for errors
	if err != nil {
		log.Println(err)
		return nil
	}

	return conn
}

// Sends a json rpc command to threw the socket and return the answer
func sendCommand(conn *net.Conn, cmd string) string {
	//Write the command to the socket
	fmt.Fprintf(*conn, cmd)
	//Read the response
	status, err := bufio.NewReader(*conn).ReadString('\n')
	//Check for any errors
	if err != nil {
		//Check for errors
		if err == io.EOF {
			/*
			 * Cgminer sends out EOF after each call.
			 * Catch this error because it's not really
			 * an error that crash the program.
			 */

		} else {
			//If the error is not EOF then warn about it
			log.Println("Sending command error: ", err)
		}
	}
	//Return the status we got from the server
	return status
}

/*
 * Bellow here is only structs defined
 * for converting json responces to
 * structs.
 */

//{"command":"summary"}
type SummaryResponse struct {
	//Status  []StatusObject  `json:"STATUS"`
	Status  []StatusObject  `json:"STATUS"`
	Summary []SummaryObject `json:"SUMMARY"`
	Id      int             `json:"id"`
}

type StatusObject struct {
	Status      string `json:"STATUS"`
	When        int    `json:"When"`
	Code        int    `json:"Code"`
	Msg         string `json:"Msg"`
	Description string `json:"Description"`
}

type SummaryObject struct {
	Elapsed            int     `json:"Elapsed"`
	MHSAv              float64 `json:"MHS av"`
	FoundBlocks        int     `json:"Found blocks"`
	Getworks           int     `json:"Getworks"`
	Accepted           int     `json:"Accepted"`
	Rejected           int     `json:"Rejected"`
	HardwareErrors     int     `json:"Hardware Errors"`
	Utility            float64 `json:"Utility"`
	Discarded          int     `json:"Discarded"`
	Stale              int     `json:"Stale"`
	GetFailures        int     `json:"Get Failures"`
	LocalWork          int     `json:"Local Work"`
	RemoteFailures     int     `json:"Remote Failures"`
	NetworkBlocks      int     `json:"Network Blocks"`
	TotalMH            float64 `json:"TotalMH"`
	WorkUtility        float64 `json:"Work Utility"`
	DifficultyAccepted float64 `json:"Difficulty Accepted"`
	DifficultyRejected float64 `json:"Difficulty Rejected"`
	DifficultyStale    float64 `json:"Difficulty Stale"`
	BestShare          int     `json:"Best Share"`
}
