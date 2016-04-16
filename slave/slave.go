package main

import (
	"time"

	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
)



func Run() {
	innerChan, outerChan, floorChan := driver.InitElevator()
	slaveSend := make(chan network.Message)
	masterSend := make(chan network.Message)
	nw := network.InitNetwork(slaveSend, masterSend)
	slaveReceive := nw.SChannels()
	go master.Run(nw, masterSend)
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	masterID := cl.Unknown
	startupTimer := time.NewTimer(50 * time.Millisecond)
	slaveSend <- network.Message{
		Receiver: nw.LocalIP,
		ID:       network.CreateID(cl.Slave),
		Response: cl.Startup,
		Content:  time.Now(),
	}
	for {
		select {
		case innerOrder := <-innerChan:
			message := network.Message{
				Receiver: masterID,
				ID:       network.CreateID(cl.Slave),
				Response: cl.InnerOrder,
				Content:  innerOrder,
			}
			slaveSend <- message
		case outerOrder := <-outerChan:
			message := network.Message{
				Receiver: masterID,
				ID:       network.CreateID(cl.Slave),
				Response: cl.OuterOrder,
				Content:  outerOrder,
			}
			slaveSend <- message
		case newFloor := <-floorChan:
			message := network.Message{
				Receiver: masterID,
				ID:       network.CreateID(cl.Slave),
				Response: cl.Floor,
				Content:  newFloor,
			}
			slaveSend <- message
		case <-doorTimer.C:
			driver.SetDoorLamp(0)
			slaveSend <- network.Message{
				Receiver: masterID,
				ID:       network.CreateID(cl.Slave),
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
				Receiver: nw.LocalIP,
				ID:       network.CreateID(cl.Slave),
				Response: cl.SetMaster,
				Content:  time.Now(),
			}
			masterID = nw.LocalIP
		}
	}
}

func main() {
	Run()
}
