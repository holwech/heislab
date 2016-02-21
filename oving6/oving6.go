package main

import (
	"net"
	"os"
	"time"
	"fmt"
	"strings"
	"math/rand"
	"os/exec"
)


func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func checkAlive(quit <-chan bool) <-chan bool{

	timeout := make(chan bool)
	go func(){
		address := "localhost:28000"
		udpAddr, err := net.ResolveUDPAddr("udp4", address)
		checkError(err)
		conn, err := net.ListenUDP("udp", udpAddr)
		defer conn.Close()
		checkError(err)
		conn.SetReadDeadline(time.Now().Add(1*time.Second))
		
		var buf [8]byte
		for{
			select{
			case <- quit:
				close(timeout)
				return
			default:
				n, _, err := conn.ReadFromUDP(buf[0:])
				if nerr, ok := err.(net.Error); ok && nerr.Timeout(){
					select {
						case timeout <- true:
						default:
					}
				}else if strings.Compare("alive",string(buf[0:n])) == 0{	
						conn.SetReadDeadline(time.Now().Add(1*time.Second))
				}
			}
		}
	}()
	return timeout
}

func pingAlive(quit <-chan bool){
	service := "localhost:28000"
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	checkError(err)

	for{
		select{
		case <-quit:
			return
		default:
			_, err = conn.Write([]byte("alive"))
			checkError(err)
			time.Sleep(10*time.Millisecond)	
		}
	}
}	


func primary(){
	for i := 1;	i < 1000; i++{
		fmt.Println(i)
		time.Sleep(1*time.Second)
	}
}

func randomExit(duration int){
	rand.Seed(time.Now().UTC().UnixNano())

	for{
		i := rand.Int()
		if (i) % (duration) == 0{
			fmt.Println("Random exit")
			os.Exit(1)
		}
	
		time.Sleep(1*time.Second)
	}
}

func main(){
	quitCheck := make(chan bool)
	timeout := checkAlive(quitCheck)

	quitCheck <-<- timeout

	quitPing := make(chan bool)
	go pingAlive(quitPing)
	go primary()

    cmd := exec.Command("bash", "-c", "start go run oving6.go")
    cmd.Start()

    neverExit := make(chan bool)
    <-neverExit
}
