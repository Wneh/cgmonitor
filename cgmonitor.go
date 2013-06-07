package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {

	conn, err := net.Dial("tcp", "192.168.1.102:4028")

	if err != nil {
		panic(err)
	}
	fmt.Fprintf(conn, "{\"command\":\"version\"}")
	status, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Println(status)
}
