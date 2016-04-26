package network

import (
	"testing"
	"time"
	"fmt"
	"github.com/holwech/heislab/cl"
)

const m = true
const s = true

func TestSend(t *testing.T) {
	nw, _ := InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	receive, send := nw.Channels()
	go senderNetwork(send)
	time.Sleep(time.Second)
	go receiverNetwork(receive)
	time.Sleep(time.Second * 360)
}

func TestListen(t *testing.T) {
	nw, _ := InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	receive, _ := nw.Channels()
	time.Sleep(time.Second)
	go receiverNetwork(receive)
	time.Sleep(time.Second * 360)
}


func senderNetwork(send chan<- Message) {
	count := 0
	for{
		time.Sleep(time.Second * 5)
		message := Message{
			LocalIP(),
      LocalIP(),
			CreateID(cl.Slave),
			"Test",
			count,
		}
		count += 1
		send <- message
		fmt.Printf("DEBUG: Sending %d\n", count)
		PrintMessage(&message)
	}
}


func receiverNetwork(receive <- chan Message) {
	for {
		select{
		case <- receive:
		}
	}
}
