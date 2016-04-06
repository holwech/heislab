package network

import (
	"testing"
	"time"
	"github.com/fatih/color"
)

func TestSend(t *testing.T) {
	slaveSend := make(chan Message)
	masterSend := make(chan Message)
	nw := new(Network)
	nw.Init(slaveSend, masterSend)
	Run(nw)
	slaveReceive, slaveStatus := nw.SChannels()
	masterReceive, masterStatus := nw.MChannels()
	go sender(slaveSend)
	time.Sleep(time.Second)
	go receiver(slaveReceive, slaveStatus, masterReceive, masterStatus)
	time.Sleep(time.Second * 60)
}

func sender(slaveSend chan Message) {
	count := 0
	for{
		time.Sleep(time.Second * 5)
		id, _ := CreateID("Slave")
		message := Message{
			LocalIP(),
			LocalIP(),
			id,
			"Test",
			count,
		}
		count += 1
		slaveSend <- message
		color.Red("DEBUG: Sending %d\n", count)
	}
}


func receiver(slaveReceive <- chan Message, slaveStatus <- chan Message, masterReceive <- chan Message, masterStatus <- chan Message) {
	for {
		select{
		case message := <- slaveReceive:
				color.Blue("Message to slave")
				PrintMessage(&message)
		case status := <- slaveStatus:
				color.Blue("Status to slave")
				PrintMessage(&status)
		case message := <- masterReceive:
				color.Blue("Message to master")
				PrintMessage(&message)
		case status := <- masterStatus:
				color.Blue("Status to master")
				PrintMessage(&status)
		}
	}
}
