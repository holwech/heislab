package backup

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

const port = ":25000"

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err)
	}
}

func pingAlive() {
	broadcastAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1"+port)
	checkError(err)
	connection, err := net.DialUDP("udp", nil, broadcastAddress)
	checkError(err)
	for {
		_, err = connection.Write([]byte("alive"))
		time.Sleep(time.Second)
	}
}

func listen() {
	timeout := time.NewTimer(3 * time.Second)
	input := make(chan string)
	quit := false
	localAddress, err := net.ResolveUDPAddr("udp", "localhost"+port)
	checkError(err)
	connection, err := net.ListenUDP("udp", localAddress)
	checkError(err)
	defer connection.Close()
	fmt.Println("BACKUP")
	go func(input chan string, connection *net.UDPConn) {
		for {
			buffer := make([]byte, 4096)
			length, _, _ := connection.ReadFromUDP(buffer)
			buffer = buffer[:length]
			input <- string(buffer)
		}
	}(input, connection)
	for {
		select {
		case <-timeout.C:
			fmt.Println("Crash detected, starting new process")
			quit = true
		case message := <-input:
			if message == "alive" {
				timeout.Reset(2 * time.Second)
			}
		}
		if quit {
			break
		}
	}
}

func Run(flag string) {
	if flag == "-b" {
		listen()
		cmd := exec.Command("bash", "-c", "gnome-terminal -x go run main.go -b")
		cmd.Start()
		time.Sleep(time.Second)
		go pingAlive()
	} else {
		cmd := exec.Command("bash", "-c", "gnome-terminal -x go run main.go -b")
		cmd.Start()
		time.Sleep(time.Second)
		go pingAlive()
	}
}
