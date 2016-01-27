/*
package main

import (
	"github.com/holwech/heislab/udp"
	"fmt"
)

func main(){
	sChan := make(chan udp.Udp_message)
	rChan := make(chan udp.Udp_message)
	err := udp.Udp_init(9000, 30000, 1024, sChan, rChan)
	strtemp := <- rChan.Data
	fmt.Println(strtemp)

}
*/

package main 

import (
	"fmt"
	"net"
	"time"
)

func udp_receive(rChan chan []byte){
	baddr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+"30000")
	broadcastListenConn, _ := net.ListenUDP("udp", baddr)

	for{
		buf := make([]byte, 1024)
		broadcastListenConn.ReadFromUDP(buf)
		fmt.Println(string(buf))
	}
	
}

func udp_send(){
	baddr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:"+"20003")
	broadcastConnection, _ := net.ListenUDP("udp", baddr)

	for {
		broadcastConnection.WriteToUDP([]byte("foollol"), baddr)
		time.Sleep(1*time.Second) 	
	}
}


func main (){
	rChan := make(chan []byte)
	go udp_send()
	udp_receive(rChan);
	for{
		select {
			case buf := <- rChan

		}
	}
}