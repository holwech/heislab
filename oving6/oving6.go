package main

import (
	"net"
	"os"
	"time"
	"fmt"
)


func udpClient(){
	service := "localhost:28000"

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	for{
		_, err = conn.Write([]byte("Client message"))
		checkError(err)

		var buf [1024]byte
		n, err := conn.Read(buf[0:])
		checkError(err)

		fmt.Println("Msg received in udpClient:", string(buf[0:n]))
		time.Sleep(1*time.Second)
	}
}

func udpServer(){
	service := "localhost:28000"
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	for {

		var buf [1024]byte
		length, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println("error reading from udp")
		}

		fmt.Println("Msg received in udpServer: ",string(buf[0:length]))

		reply := "Server msg"
		conn.WriteToUDP([]byte(reply), addr)
	}
}

func main() {
	go udpServer()
	time.Sleep(1*time.Second)
	go udpClient()

	neverStop := make(chan int)
	<- neverStop
}

