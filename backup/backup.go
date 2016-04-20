package backup

import (
	"net"
	"os"
	"time"
	"fmt"
	"os/exec"
)

const local_IP = "localhost"
const port = ":25051"

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err)
		os.Exit(1)
	}
}

func pingAlive(){
	broadcastAddress, err := net.ResolveUDPAddr("udp", local_IP + port)
	checkError(err)
	localAddress, err := net.ResolveUDPAddr("udp", local_IP + port)
	checkError(err)
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	checkError(err)
	_, err = connection.Write([]byte("quit"))
	defer connection.Close()
	for {
		_, err = connection.Write([]byte("alive"))
		checkError(err)
		fmt.Println("Alive")
		time.Sleep(time.Second)
	}
}

func listen() (<- chan time.Time) {
	timeout := time.NewTimer(2 * time.Second)
	go func(){
		localAddress, err := net.ResolveUDPAddr("udp", port)
		checkError(err)
		connection, err := net.ListenUDP("udp", localAddress)
		checkError(err)
		defer connection.Close()
		for {
			buffer := make([]byte, 4096)
			fmt.Println("Before Read")
			length, _, err := connection.ReadFromUDP(buffer)
			checkError(err)
			fmt.Println("After read")
			buffer = buffer[:length]
			if string(buffer) == "alive" {
				timeout.Reset(2 * time.Second)
				fmt.Println("Alive received")
			} else if string(buffer) == "quit" {
				break
			}
		}
	}()
	return timeout.C
}


func Run(flag string) {
	if flag == "-b" {
		timeout := listen()
		fmt.Println("Waiting")
		<- timeout
		fmt.Println("Done waiting")
		go pingAlive()
		time.Sleep(100 * time.Millisecond)
		cmd := exec.Command("bash", "-c", "gnome-terminal -x go run main.go")
		cmd.Start()
	} else {
		go pingAlive()
		cmd := exec.Command("bash", "-c", "gnome-terminal -x go run main.go")
		cmd.Start()
	}
}
