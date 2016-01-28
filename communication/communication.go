package communication

import (
	"time"
	"net"
	"fmt"
)

func Broadcast() {
	fmt.Println("=== Broadcasting ===")
	broadcastAddress, _ := net.ResolveUDPAddr("udp4", "129.241.187.255:20000")
	localAddress := getLocalAddress(broadcastAddress)
	connection, _ := net.DialUDP("udp4",localAddress, broadcastAddress)
	defer connection.Close()
	for {
		connection.Write([]byte("Penis?"))
		time.Sleep(1*time.Second)
	}
}

func Listen(InputUDP chan string) {
	fmt.Println("=== Listening ===")
	broadcastAddress, _ := net.ResolveUDPAddr("udp4", "129.241.187.255:20001")
	localAddress := getLocalAddress(broadcastAddress)
	connection, _ := net.ListenUDP("udp4", localAddress)
	defer connection.Close()
	for{
		buffer := make([]byte, 1024)
		connection.ReadFromUDP(buffer)
		InputUDP <- string(buffer)
	}
}

func getLocalAddress(receiveAddress *net.UDPAddr) (*net.	UDPAddr) {
	tempConnection, _ := net.DialUDP("udp4", nil, receiveAddress)
	defer tempConnection.Close()
	tempAddress := tempConnection.LocalAddr()
	localAddress, _ := net.ResolveUDPAddr("udp4", tempAddress.String())
	localAddress.Port = 20000
	return localAddress
}