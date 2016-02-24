package main

import (
	"net"
	"os"
	"time"
	"fmt"
	"math/rand"
	"os/exec"
	"encoding/binary"
)


func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err)
		os.Exit(1)
	}
}

func checkPrimary(quit <-chan bool) (<-chan int){
	timeout := make(chan int)
	
	go func(){
		address := "127.0.0.1:25000"
		udpAddr, err := net.ResolveUDPAddr("udp4", address)
		checkError(err)
		conn, err := net.ListenUDP("udp4", udpAddr)
		defer conn.Close()
		checkError(err)
		conn.SetReadDeadline(time.Now().Add(1*time.Second))
		
		prevVal := 0
		buf := make([]byte, 8)
		for{
			select{
			case <- quit:
				close(timeout)
				return
			default:
				_, _, err := conn.ReadFromUDP(buf[0:])
				if nerr, ok := err.(net.Error); ok && nerr.Timeout(){
					select {
						case timeout <- prevVal:
						default:
					}
				}else{
				 	recvVal := int(binary.BigEndian.Uint64(buf)) 
				 	prevVal = recvVal
					conn.SetReadDeadline(time.Now().Add(1*time.Second))		
				}
			}
		}
	}()
	return timeout
}

func pingAlive(pingVal <-chan int, quit <-chan bool){
	service := "127.0.0.1:25000"
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	defer conn.Close()
	checkError(err)
	ping := 0

	for{
		select{
		case <-quit:
			return
		case ping = <-pingVal:
		default:
			buf := make([]byte,10)
		    binary.BigEndian.PutUint64(buf, uint64(ping))
			_, err = conn.Write(buf)
			checkError(err)
			time.Sleep(10*time.Millisecond)	
		}
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
	timeout := checkPrimary(quitCheck)

	currentVal := <- timeout
	quitCheck <- true

    cmd := exec.Command("bash", "-c", "gnome-terminal -x go run oving6.go")
    cmd.Start()

    time.Sleep(1*time.Second)

	quitPing := make(chan bool)
	pingVal := make (chan int)

	go pingAlive(pingVal, quitPing)
	//go randomExit(10)


    //cmd := exec.Command("bash", "-c", "start go run oving6.go")

 	for{
		pingVal <- currentVal		
		fmt.Println(currentVal+1)
		currentVal += 1
		time.Sleep(1*time.Second)
	}
	fmt.Println("finito")
}
