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
	baddr,_ := net.ResolveUDPAddr("udp4", ":20056")
	conn, _ := net.ListenUDP("udp4", baddr)
	defer conn.Close()

	for{
		buf := make([]byte, 1024)
		conn.ReadFromUDP(buf)
		rChan <- buf
	}
}

func udp_send(){	
	ip := "129.241.187.146:"

	baddr, _ := net.ResolveUDPAddr("udp4", ip+"20056")
	conn,_ := net.DialUDP("udp",nil, baddr)
	defer conn.Close()

	for {
		conn.Write([]byte("foollol"))
		time.Sleep(1*time.Second) 	
	}
}


func main (){
	rChan := make(chan []byte)

	go udp_send()
	go udp_receive(rChan);

	for{
		fmt.Println(string(<-rChan))
	}
	
}