package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

//Struct representing the config.toml
type Config struct {
	Webserverport int              //Webserver port
	Miners        map[string]miner //Key is the miner name.
	Users         map[string]user  //Users for login
}

//Struct for config file type [miners.<foo>] 
type miner struct {
	IP        string  //The ip
	Threshold float64 //The threshold that the miner cant go below(in mega hashes)
	KeepAlive bool    //If restart command should be send on sick/dead detection
}

//Config file struct for user
type user struct {
	Username string //The username - used as key in Users map
	Salt     string //The salt used for this user
	Hash     string //Password salted with the salt and hashed
}

//"Top" struct that represent each miner.
//This struct is used in the var miners that is called from here and there
type MinerInformation struct {
	Name         string         //The miners name
	SumWrap      SummaryWrapper //Summary
	DevsWrap     DevsWrapper    //Devs
	Client       *Client        //RPC Client that holds information
	ClientConfig miner
}

//Wrapper that contain all the information that is used for the summary request
type SummaryWrapper struct {
	Mu         sync.RWMutex    //Mutex lock
	Summary    SummaryResponse //The response parsed to a struct
	SummaryRow MinerRow        //The response converted to rows for the web page
}

//Wrapper that contain all the information that is used for the devs request
type DevsWrapper struct {
	Mu   sync.RWMutex //The mutex lock
	Devs DevsResponse //The response from devs in json format 
}

//Global variabels
//Map containing with the miner name as key
var miners map[string]*MinerInformation

//Config file
var config Config

func main() {
	//Open the log file
	logf, err := os.OpenFile("cgmonitor.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
	if err != nil {
		log.Fatal(err)
	}
	//Make sure that we close it later on
	defer logf.Close()

	//Set log to both file and console
	log.SetOutput(io.MultiWriter(logf, os.Stdout))

	log.Println("Begin reading config file...")
	//Check that config file exists
	configExists()
	//Read the config file
	config.Load()
	log.Println("...done reading config file")

	//Check if the user wants to add a new user for the web
	if len(os.Args) > 1 {
		if os.Args[1] == "--addNewUser" {
			addNewUser()
			return
		}
	}

	log.Println("Starting server...")

	//Create a waitgroup
	wg := new(sync.WaitGroup)

	miners = make(map[string]*MinerInformation)

	log.Println("Begin starting rpc-client threads...")
	//Start to grab information from every miner
	for minerName, miner := range config.Miners {
		//Add two(one for each thread) to the WaitGroup for each miner
		wg.Add(2)

		log.Printf("Starting: %s(%s)\n", minerName, miner.IP)
		//Create a new miner struct and add the name
		minerStructTemp := MinerInformation{}
		minerStructTemp.Name = minerName
		minerStructTemp.ClientConfig = miner

		//Add save it
		miners[minerName] = &minerStructTemp

		//Start one new gorutine for each miner
		go rpcClient(minerName, miner.IP, 10, &minerStructTemp, wg, miner.Threshold)
	}
	log.Println("...Waiting for every thread to be started")
	wg.Wait()

	log.Println("All thread started, starting web server")
	webServerMain(config.Webserverport)
}

//Check of the there is a config.toml file is the same folder as the program is runned from
//If not it will create one
func configExists() {
	if _, err := os.Stat("cgmonitor.conf"); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			log.Println("No config file found, creating example config file.")
			createExampleConf()
		} else {
			// other error
		}
	}
}

//Creates a basic config file
func createExampleConf() {
	//Create the Config struct and add some values
	tempConf := Config{8080, make(map[string]miner), make(map[string]user)}
	tempConf.Miners["alpha"] = miner{"127.0.0.1:4028", 0.1, true}

	//Save it to cgmonitor.conf
	tempConf.Save()
}

func addNewUser() {
	//Print warning
	fmt.Println("###################################################")
	fmt.Println("# Note:                                           #")
	fmt.Println("# ----------------------------------------------- #")
	fmt.Println("# DONT use a password that you have one any site. #")
	fmt.Println("# Since I can't guarantee that my login           #")
	fmt.Println("# implementation is secure.                       #")
	fmt.Println("# ----------------------------------------------- #")
	fmt.Println("# Confirm that you have understood by typing: yes #")
	fmt.Println("###################################################")
	var confirm string
	fmt.Scan(&confirm)
	//And check that the user have understood the warning text
	if confirm != "yes" {
		fmt.Println("CGmonitor will now exit since you didn't type: yes")
		return
	}

	//Grab input from the user
	fmt.Println("username:")
	var username string
	fmt.Scan(&username)
	fmt.Println("password:")
	var password string
	fmt.Scan(&password)

	//fmt.Printf("Username: %s \nPassword: %s\n", username, password)

	//Create the new user
	var newUser user

	newUser.Username = username
	//Generate a salt for this user
	newUser.Salt = randomString(64)
	//Hash the password with the salt
	newUser.Hash = hashPassword(password, newUser.Salt)

	//Save the user
	config.AddUser(newUser)

	log.Println("New user created:", newUser.Username)
}

//Add a user and save the config to file
func (c *Config) AddUser(newUser user) {
	if c.Users == nil {
		c.Users = make(map[string]user)
	}
	c.Users[newUser.Username] = newUser
	//Save the new config to file
	c.Save()
}

//Saves the config file to cgmonitor.conf
func (c *Config) Save() {
	//Convert it to json
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		log.Println("error:", err)
	}

	//And save it to file
	err = ioutil.WriteFile("cgmonitor.conf", b, 0644)
	if err != nil {
		panic(err)
	}
}

func (c *Config) Load() {
	b, err := ioutil.ReadFile("cgmonitor.conf")
	if err != nil {
		panic(err)
	}
	//Parse the raw json to struct
	err = json.Unmarshal(b, c)
	if err != nil {
		panic(err)
	}
}

func randomString(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func hashPassword(password, salt string) string {
	h := sha1.New()
	//fmt.Println(salt+password)
	io.WriteString(h, salt+password)
	//fmt.Printf("%x", h.Sum(nil))
	return string(hex.EncodeToString(h.Sum(nil)))
}
