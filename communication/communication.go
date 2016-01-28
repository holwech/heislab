package communication

import (
	"time"
	"net"
	"fmt"
)

func Broadcast() {
	fmt.Println("=== Broadcasting ===")
	broadcastAddress, err := net.ResolveUDPAddr("udp4", "10.20.78.255:30000")
	if err != nil {
		fmt.Println("=== ERROR: 1")
		fmt.Println(err)
	}
	localAddress := getLocalAddress(broadcastAddress)
	connection, err := net.DialUDP("udp4",localAddress, broadcastAddress)
	if err != nil {
		fmt.Println("=== ERROR: 2")
		fmt.Println(err)
	}
	defer connection.Close()
	for {
		connection.Write([]byte("Penis?"))
		time.Sleep(1*time.Second)
	}
}

func Listen(InputUDP chan string) {
	fmt.Println("=== Listening ===")
	broadcastAddress, err := net.ResolveUDPAddr("udp4", "10.20.78.108:30000")
	if err != nil {
		fmt.Println("=== ERROR: 3")
		fmt.Println(err)
	}
	localAddress := getLocalAddress(broadcastAddress)
	connection, err := net.ListenUDP("udp4", localAddress)
	if err != nil {
		fmt.Println("=== ERROR: 4")
		fmt.Println(err)
	}
	defer connection.Close()
	for{
		buffer := make([]byte, 1024)
		connection.ReadFromUDP(buffer)
		InputUDP <- string(buffer)
	}
}

func getLocalAddress(receiveAddress *net.UDPAddr) (*net.	UDPAddr) {
	tempConnection, err := net.DialUDP("udp4", nil, receiveAddress)
	defer tempConnection.Close()
	tempAddress := tempConnection.LocalAddr()
	localAddress, err := net.ResolveUDPAddr("udp4", tempAddress.String())
	if err != nil {
		fmt.Println("=== ERROR: 5")
		fmt.Println(err)
	}
	localAddress.Port = 20000
	return localAddress
}