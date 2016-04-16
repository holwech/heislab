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
	slaveReceive := nw.SChannels()
	masterReceive := nw.MChannels()
	go sender(slaveSend)
	time.Sleep(time.Second)
	go receiver(slaveReceive, masterReceive)
	time.Sleep(time.Second * 60)
}

func TestListen(t *testing.T) {
	slaveSend := make(chan Message)
	masterSend := make(chan Message)
	nw := new(Network)
	nw.Init(slaveSend, masterSend)
	Run(nw)
	slaveReceive := nw.SChannels()
	masterReceive := nw.MChannels()
	time.Sleep(time.Second)
	go receiver(slaveReceive, masterReceive)
	time.Sleep(time.Second * 60)
}


func sender(slaveSend chan Message) {
	count := 0
	for{
		time.Sleep(time.Second * 5)
		id := CreateID("Master")
		message := Message{
			LocalIP(),
			LocalIP(),
			id,
			"Test",
			count,
			//map[string]interface{}{"test": 1, "wha": "ekeke"},
		}
		count += 1
		slaveSend <- message
		color.Red("DEBUG: Sending %d\n", count)
	}
}


func receiver(slaveReceive <- chan Message, masterReceive <- chan Message) {
	for {
		select{
		case message := <- slaveReceive:
				color.Blue("Message to slave")
				PrintMessage(&message)
		case message := <- masterReceive:
				color.Blue("Message to master")
				PrintMessage(&message)
		}
	}
}
