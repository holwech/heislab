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
	nw := InitNetwork()
	slaveReceive, slaveSend := nw.SChannels()
	masterReceive, _ := nw.MChannels()
	go sender(slaveSend)
	time.Sleep(time.Second)
	go receiver(slaveReceive, masterReceive)
	time.Sleep(time.Second * 360)
}

func TestListen(t *testing.T) {
	nw := InitNetwork()
	slaveReceive, _ := nw.SChannels()
	masterReceive, _ := nw.MChannels()
	time.Sleep(time.Second)
	go receiver(slaveReceive, masterReceive)
	time.Sleep(time.Second * 360)
}


func sender(slaveSend chan<- Message) {
	count := 0
	for{
		time.Sleep(time.Second * 5)
		message := Message{
			LocalIP(),
			"129.241.187.140",
			CreateID(cl.Master),
			"Test",
			count,
		}
		count += 1
		slaveSend <- message
		fmt.Printf("DEBUG: Sending %d\n", count)
		PrintMessage(&message)
	}
}


func receiver(slaveReceive <- chan Message, masterReceive <- chan Message) {
	for {
		select{
		case message := <- slaveReceive:
				fmt.Println("Message to slave")
				PrintMessage(&message)
		case message := <- masterReceive:
				fmt.Println("Message to master")
				PrintMessage(&message)
		}
	}
}
