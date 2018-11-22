package main

import (
	"fmt"
	"os"
	"strconv"

	"net"
	"bufio"
)

func CrashOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ServerHandler(conn net.Conn) {
	fmt.Printf("New Conn From: %s\n", conn.RemoteAddr())
	fmt.Fprintf(conn, "ACK\n")
	response, err := bufio.NewReader(conn).ReadString('\n')
	CrashOnError(err)
	fmt.Println("Server Received:", response)
}

func ServerStart(port string) {
	listener, err := net.Listen("tcp", port)
	CrashOnError(err)
	for {
		conn, err := listener.Accept()
		CrashOnError(err)
		ServerHandler(conn)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: node <peerCount> <myId>")
		os.Exit(1)
	}

	peerCount, err := strconv.Atoi(os.Args[1])
	CrashOnError(err)
	myId, err :=  strconv.Atoi(os.Args[2])
	CrashOnError(err)

	fmt.Printf("Starting Blockchain node myId=%d peerCount=%d\n", myId, peerCount)

	myPort := fmt.Sprintf(":8%03d", myId)
	go ServerStart(myPort)

	conn, err := net.Dial("tcp", myPort)
	CrashOnError(err)
	fmt.Fprintf(conn, "HELLO\n")
	response, err := bufio.NewReader(conn).ReadString('\n')
	CrashOnError(err)
	fmt.Println("Client Received:", response)
}