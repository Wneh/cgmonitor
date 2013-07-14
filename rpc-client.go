package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	Name             string            //Name of the miner
	IP               string            //Ip to the cgminer including port
	Conn             net.Conn          //Connection made with net.Dial
	RefreshInterval  int               //Seconds between fetching information
	MinerInfo        *MinerInformation //Struct to put the answers for the webserver
	ClientRequests   chan RpcRequest   //Channel for sending rpc request to cgminer
	MHSThresLimit    float64           //MHSAv should not be below this
	LastSumTimestamp int               //Last timestamp from summary request
}

//Main function for fetching information from one client
func rpcClient(name, ip string, refInt int, minerInfo *MinerInformation, wg *sync.WaitGroup, threshold float64) {
	//Add everything except the connection
	c := Client{name, ip, nil, refInt, minerInfo, nil, threshold, int(time.Now().Unix())}
	//Save the Client struct in the MinerInfo
	c.MinerInfo.Client = &c

	clientRequests := make(chan RpcRequest)
	c.ClientRequests = clientRequests

	//Start the thread the will keep doing summary requests
	go SummaryHandler(clientRequests, minerInfo, &c, wg)
	//Start another thread the will ask the devs requests
	go DevsHandler(clientRequests, minerInfo, &c, wg)

	//Wait for new requst to make from the clienReequest channel
	for r := range clientRequests {
		//Create a new connection
		c.Conn = createConnection(c.IP)

		//If c.Conn is still nil then we couldn't connect
		//So send back an empty slice of bytes
		if c.Conn == nil {
			r.ResultChan <- make([]byte, 0)
		} else {
			//Send the request to the cgminer
			b := sendCommand(&c.Conn, r.Request)
			/* 
			 * Note:
			 *
			 * It seems that cgminer close the tcp connection
			 * after each call so we need to reset it for
			 * the next rpc-call
			 */
			c.Conn.Close()
			//And send back the result
			r.ResultChan <- b
		}
	}
}

//Making summary requests to the cgminer and parse the result.
func SummaryHandler(res chan<- RpcRequest, minerInfo *MinerInformation, c *Client, wg *sync.WaitGroup) {
	request := RpcRequest{"{\"command\":\"summary\"}", make(chan []byte)}

	var response []byte
	//Creating an empty instance of everything
	summary := SummaryResponse{[]StatusObject{StatusObject{}}, []SummaryObject{SummaryObject{}}, 0}
	summaryRow := MinerRow{}
	summaryRow.Name = c.Name

	//Save the default values
	//Lock it
	minerInfo.SumWrap.Mu.Lock()
	//Save the summary
	minerInfo.SumWrap.Summary = summary
	minerInfo.SumWrap.SummaryRow = summaryRow
	//Now unlock
	minerInfo.SumWrap.Mu.Unlock()

	//Signal that the thread is started
	wg.Done()

	for {
		res <- request
		response = <-request.ResultChan

		//If we got the response back unmarshal it
		if len(response) != 0 {
			err := json.Unmarshal(response, &summary)
			//Check for errors
			if err != nil {
				fmt.Println(err.Error())
			}

			//Update the summaryrow
			summaryRow = MinerRow{c.Name, summary.Summary[0].Accepted, summary.Summary[0].Rejected, summary.Summary[0].MHSAv, summary.Summary[0].BestShare}
		}
		//Lock it
		minerInfo.SumWrap.Mu.Lock()
		//Save the summary
		minerInfo.SumWrap.Summary = summary
		minerInfo.SumWrap.SummaryRow = summaryRow
		//Now unlock
		minerInfo.SumWrap.Mu.Unlock()

		//Now sleep
		time.Sleep(time.Duration(c.RefreshInterval) * time.Second)
	}
}

//Checks the current mhs average value against the threshold.
//The value should have been lower for 10 minutes 
//before it restarts the miner  
func CheckMhsThresHold(mhs float64, lasttime int, c *Client) {
	switch {
	//Good - It's abowe the limit
	case mhs >= c.MHSThresLimit:
		//Save the last timestamp
		c.LastSumTimestamp = lasttime
		fmt.Printf("Hashrate: Good(%v > %v)\n", mhs, c.MHSThresLimit)
		return
	//Meeh - Under the limit but it hasn't gone 10 min yey
	case mhs < c.MHSThresLimit && (lasttime-c.LastSumTimestamp) < 600:
		//Dont to nothing just wait and see if the hashrate
		//goes up or if it keeps down
		fmt.Printf("Hashrate: Below threshold(%v < %v) for %v secs which is under 10 min\n", mhs, c.MHSThresLimit, (lasttime - c.LastSumTimestamp))
		return
	//Oh noes - Below the threshold and for longer then 10 min
	case mhs < c.MHSThresLimit && (lasttime-c.LastSumTimestamp) >= 600:
		//Restart the miner
		fmt.Printf("Hashrate: Below threshold(%v < %v) for %v secs which is over 10 min\n", mhs, c.MHSThresLimit, (lasttime - c.LastSumTimestamp))
		return
	}
}

//Making devs request to the cgminer and parse the result
func DevsHandler(res chan<- RpcRequest, minerInfo *MinerInformation, c *Client, wg *sync.WaitGroup) {
	request := RpcRequest{"{\"command\":\"devs\"}", make(chan []byte)}

	var response []byte
	var devs DevsResponse

	//Signal that the thread is started
	wg.Done()

	for {
		res <- request
		response = <-request.ResultChan

		if len(response) != 0 {
			err := json.Unmarshal(response, &devs)
			//Check for errors
			if err != nil {
				fmt.Println(err.Error())
			}
			mhs5s := 0.0
			//Set the onoff boolean for every device
			//Also sum up the MHS5s
			for i := 0; i < len(devs.Devs); i++ {
				//Get the variable
				var dev = &devs.Devs[i]
				//Make the comparison
				if dev.Enabled == "Y" {
					dev.OnOff = true
				} else {
					dev.OnOff = false
				}
				//Sum upp the MHS5s
				mhs5s += dev.MHS5s
			}
			CheckMhsThresHold(mhs5s, devs.Status[0].When, c)
		}

		//Lock it
		minerInfo.DevsWrap.Mu.Lock()
		//Save the summary
		minerInfo.DevsWrap.Devs = devs
		//Now unlock
		minerInfo.DevsWrap.Mu.Unlock()

		//Now sleep
		time.Sleep(time.Duration(c.RefreshInterval) * time.Second)
	}
}

func enableDisable(status, device int, name string) {
	var request RpcRequest

	switch status {
	case 0:
		request = RpcRequest{fmt.Sprintf("{\"command\":\"gpudisable\",\"parameter\":\"%v\"}", device), make(chan []byte)}
	case 1:
		request = RpcRequest{fmt.Sprintf("{\"command\":\"gpuenable\",\"parameter\":\"%v\"}", device), make(chan []byte)}
	}

	fmt.Println("The request:", request.Request)

	var response []byte
	//var devs DevsResponse

	miners[name].Client.ClientRequests <- request
	response = <-request.ResultChan

	fmt.Println("Result from restart:", response)
}

// Returns a TCP connection to the ip 
func createConnection(ip string) net.Conn {
	conn, err := net.Dial("tcp", ip)

	//Check for errors
	if err != nil {
		log.Printf("createConnection: %s, check if the ip is correct or cgminer's api is enabled", err)
		return nil
	}
	return conn
}

// Sends a json rpc command to threw the socket and return the answer
func sendCommand(conn *net.Conn, cmd string) []byte {
	//Write the command to the socket
	fmt.Fprintf(*conn, cmd)
	//Read the response
	response, err := bufio.NewReader(*conn).ReadString('\n')
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
	//Create the byte array
	b := []byte(response)

	/*
	 * Check for \x00 to remove
	 */
	if b[len(b)-1] == '\x00' {
		b = b[0 : len(b)-1]
	}

	//Return the status we got from the server
	return b
}

type RpcRequest struct {
	Request    string
	ResultChan chan []byte
}

/*
 * Bellow here is only structs defined
 * for converting json responces to
 * structs.
 */

////////////
// Status //
////////////
type StatusObject struct {
	Status      string `json:"STATUS"`
	When        int    `json:"When"`
	Code        int    `json:"Code"`
	Msg         string `json:"Msg"`
	Description string `json:"Description"`
}

/////////////
// summary //
/////////////
type SummaryResponse struct {
	Status  []StatusObject  `json:"STATUS"`
	Summary []SummaryObject `json:"SUMMARY"`
	Id      int             `json:"id"`
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

//////////
// devs //
//////////
type DevsResponse struct {
	Status []StatusObject `json:"STATUS"`
	Devs   []DevObject    `json:"DEVS"`
	Id     int            `json:"id"`
}

type DevObject struct {
	GPU                 int     `json:"GPU"`
	Enabled             string  `json:"Enabled"`
	Status              string  `json:"Status"`
	Temperature         float64 `json:"Temperature"`
	FanSpeed            int     `json:"Fan Speed"`
	FanPercent          int     `json:"Fan Percent"`
	GPUClock            int     `json:"GPU Clock"`
	MemoryClock         int     `json:"Memory Clock"`
	GPUVoltage          float64 `json:"GPU Voltage"`
	GPUActivity         int     `json:"GPU Activity"`
	Powertune           int     `json:"Powertune"`
	MHSAv               float64 `json:"MHS av"`
	MHS5s               float64 `json:"MHS 5s"`
	Accepted            int     `json:"Accepted"`
	Rejected            int     `json:"Rejected"`
	HardwareErrors      int     `json:"Hardware Errors"`
	Utility             float64 `json:"Utility"`
	Intensity           string  `json:"Intensity"`
	LastSharePool       int     `json:"Last Share Pool"`
	LastShareTime       int     `json:"Last Share Time"`
	TotalMH             float64 `json:"Total MH"`
	Diff1Work           int     `json:"Diff1 Work"`
	DifficultyAccepted  float64 `json:"Difficulty Accepted"`
	DifficultyRejected  float64 `json:"Difficulty Rejected"`
	LastShareDifficulty float64 `json:"Last Share Difficulty"`
	LastValidWork       int     `json:"Last Valid Work"`
	OnOff               bool    //This is an extra boolean used for html template parsing
}
