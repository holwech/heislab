package network

import (
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	slaveSend := make(chan Message)
	masterSend := make(chan Message)
	nw := new(Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)
	slaveReceive, slaveStatus := nw.SChannels()
	go sender(slaveSend)
	time.Sleep(time.Second)
	go receiver(slaveReceive)
	time.Sleep(time.Second * 60)
}

func sender(slaveSend chan Message) {
	count := 0
	for{
		time.Sleep(time.Second * 5)
		message := network.Message{
			network.LocalIP(),
			network.LocalIP(),
			network.CreateID("Slave"),
			"Test",
			count,
		}
		count += 1
		slaveSend <- message
	}
}


func receiver(slaveReceive chan Message) {
	for {
		message := <- slaveReceive
		PrintMessage(&message)
	}
}
