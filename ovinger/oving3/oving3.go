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
//	"time"
//	"github.com/holwech/heislab/tcp"
//	"strconv"
//	"flag"
)

// func udp_receive(rChan chan []byte){
// 	baddr,_ := net.ResolveUDPAddr("udp4", "129.241.187.255:30000")
// 	conn, _ := net.ListenUDP("udp4", baddr)
// 	defer conn.Close()

// 	for{
// 		buf := make([]byte, 1024)
// 		conn.ReadFromUDP(buf)
// 		rChan <- buf
// 	}
// }

// func udp_send(){	
// 	ip := "129.241.187.146:"

// 	baddr, _ := net.ResolveUDPAddr("udp4", ip+"20056")
// 	conn,_ := net.DialUDP("udp",nil, baddr)
// 	defer conn.Close()

// 	for {
// 		conn.Write([]byte("foollol"))
// 		time.Sleep(1*time.Second) 	
// 	}
// }





func main (){
	rAddr, _ := net.ResolveTCPAddr("tcp4", "129.241.187.23:33546")
	lAddr, _ := net.ResolveTCPAddr("tcp4", "129.241.187.146:33546")	
	conn, _ := net.DialTCP("tcp", lAddr, rAddr)
	defer conn.Close()
	conn.Write([]byte("Hei penis \x00"))
	for{
		buf := make([]byte, 1024)
		conn.Read(buf)
		fmt.Println(string(buf))
	}

}