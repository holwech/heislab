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
	go sender(send)
	time.Sleep(time.Second)
	go receiver(receive)
	time.Sleep(time.Second * 360)
}

func TestListen(t *testing.T) {
	nw, _ := InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	receive, _ := nw.Channels()
	time.Sleep(time.Second)
	go receiver(receive)
	time.Sleep(time.Second * 360)
}


func sender(send chan<- Message) {
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


func receiver(receive <- chan Message) {
	for {
		select{
		case <- receive:
		}
	}
}
