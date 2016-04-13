package main

import (
	"time"

	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
)

func InitElevator() (<-chan driver.InnerOrder, <-chan driver.OuterOrder, <-chan int) {
	driver.InitHardware()
	innerChan := driver.ListenInnerPanel()
	outerChan := driver.ListenOuterPanel()
	floorChan := driver.ListenFloorSensor()

	//Drive down to first floor
	currentFloor := <-floorChan
	if currentFloor != 0 {
		driver.SetMotorDirection(-1)
		for currentFloor != 0 {
			currentFloor = <-floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan, outerChan, floorChan
}

func InitNetwork(slaveSend chan network.Message, masterSend chan network.Message) *network.Network {
	nw := new(network.Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)
	return nw
}

func Run() {
	innerChan, outerChan, floorChan := InitElevator()
	slaveSend := make(chan network.Message)
	masterSend := make(chan network.Message)
	nw := InitNetwork(slaveSend, masterSend)
	slaveReceive, slaveStatus := nw.SChannels()
	go master.Run(nw, masterSend)
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	masterID := cl.Unknown
	startupTimer := time.NewTimer(50 * time.Millisecond)
	slaveSend <- network.Message{
		Sender:   network.LocalIP(),
		Receiver: network.LocalIP(),
		ID:       network.CreateID("Slave"),
		Response: cl.Startup,
		Content:  time.Now(),
	}
	for {
		select {
		case innerOrder := <-innerChan:
			message := network.Message{
				Sender:   network.LocalIP(),
				Receiver: masterID,
				ID:       network.CreateID("Slave"),
				Response: cl.InnerOrder,
				Content:  innerOrder,
			}
			slaveSend <- message
		case outerOrder := <-outerChan:
			message := network.Message{
				Sender:   network.LocalIP(),
				Receiver: masterID,
				ID:       network.CreateID("Slave"),
				Response: cl.OuterOrder,
				Content:  outerOrder,
			}
			slaveSend <- message
		case newFloor := <-floorChan:
			message := network.Message{
				Sender:   network.LocalIP(),
				Receiver: masterID,
				ID:       network.CreateID("Slave"),
				Response: cl.Floor,
				Content:  newFloor,
			}
			slaveSend <- message
		case <-doorTimer.C:
			driver.SetDoorLamp(0)
			slaveSend <- network.Message{
				Sender:   network.LocalIP(),
				Receiver: masterID,
				ID:       network.CreateID("Slave"),
				Response: cl.DoorClosed,
				Content:  "",
			}
		case message := <-slaveReceive:
			switch message.Response {
			case cl.Up:
				driver.SetMotorDirection(1)
			case cl.Down:
				driver.SetMotorDirection(-1)
			case cl.Stop:
				driver.SetMotorDirection(0)
				driver.SetDoorLamp(1)
				doorTimer.Reset(3 * time.Second)
			case cl.JoinMaster:
				startupTimer.Stop()
				masterID = message.Sender
			case cl.LightOnInner:
				driver.SetInnerPanelLamp(int(message.Content.(float64)), 1)
			case cl.LightOffInner:
				driver.SetInnerPanelLamp(int(message.Content.(float64)), 0)
			case cl.LightOnOuterUp:
				driver.SetOuterPanelLamp(1, int(message.Content.(float64)), 1)
			case cl.LightOffOuterUp:
				driver.SetOuterPanelLamp(1, int(message.Content.(float64)), 0)
			case cl.LightOnOuterDown:
				driver.SetOuterPanelLamp(-1, int(message.Content.(float64)), 1)
			case cl.LightOffOuterDown:
				driver.SetOuterPanelLamp(-1, int(message.Content.(float64)), 0)
			}
		case <-startupTimer.C:
			slaveSend <- network.Message{
				Sender:   network.LocalIP(),
				Receiver: network.LocalIP(),
				ID:       network.CreateID("Slave"),
				Response: cl.SetMaster,
				Content:  time.Now(),
			}
			masterID = nw.LocalIP
		case <-slaveStatus:
			break
		}
	}
}

func main() {
	Run()
}
