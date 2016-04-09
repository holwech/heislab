package main

import (
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/master"
	"time"
	"fmt"
)

func InitElevator() (<-chan driver.InnerOrder,<-chan driver.OuterOrder, <-chan int){
	driver.InitHardware()
	
	innerChan := driver.ListenInnerPanel()
	outerChan := driver.ListenOuterPanel()
	floorChan := driver.ListenFloorSensor()

	//Drive down to first floor
	currentFloor := <-floorChan
	if currentFloor != 0{
		driver.SetMotorDirection(-1)
		for currentFloor != 0{
			currentFloor = <-floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan,outerChan,floorChan
}

func InitNetwork(slaveSend chan network.Message, masterSend chan network.Message) *network.Network {
	nw := new(network.Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)
	return nw
}

func initSlave(slaveSend chan<- network.Message, slaveReceive <-chan network.Message) string {
	time.Sleep(time.Millisecond * 10)
	masterID := network.LocalIP()
	timer := time.NewTimer(50 * time.Millisecond).C
	msg := network.Message{
		Sender: network.LocalIP(),
		Receiver: network.LocalIP(),
		ID: network.CreateID("Slave"),
		Response: cl.Startup,
		Content: time.Now(),
	}
	slaveSend <- msg

	for{
		flag := false
		select{
		case message := <- slaveReceive:

			if message.Response == cl.JoinMaster {

				masterID = message.Sender
				flag = true
				fmt.Println("Master found")
			}
		case <- timer:

			flag = true
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: network.LocalIP(),
				ID: network.CreateID("Slave"),
				Response: cl.SetMaster,
				Content: time.Now(),
			}
			slaveSend <- message
			fmt.Println("No master found")
		}
		if flag == true {
			break
		}
	}

	return masterID
}

func Run() {
	innerChan, outerChan, floorChan := InitElevator()
	slaveSend := make(chan network.Message)
	masterSend := make(chan network.Message)
	nw := InitNetwork(slaveSend, masterSend)
	slaveReceive, slaveStatus := nw.SChannels()
	go master.Run(nw, masterSend)
	masterID := initSlave(slaveSend, slaveReceive)
	doorTimer := time.NewTimer(3 * time.Second)
	doorTimer.Stop()
	for {
		select{
		case innerOrder := <- innerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: masterID,
				ID: network.CreateID("Slave"),
				Response: cl.InnerOrder,
				Content: innerOrder,
			}
			slaveSend <- message
		case outerOrder := <- outerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: masterID,
				ID: network.CreateID("Slave"),
				Response: cl.OuterOrder,
				Content: outerOrder,
			}
			slaveSend <- message
		case newFloor := <- floorChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: masterID,
				ID: network.CreateID("Slave"),
				Response: cl.Floor,
				Content: newFloor,
			}
			slaveSend <- message
		case <- doorTimer.C:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: masterID,
				ID: network.CreateID("Slave"),
				Response: cl.DoorClosed,
				Content: "",
			}
			slaveSend <- message
		case message := <- slaveReceive:
			switch message.Response{
			case cl.Up:
				driver.SetMotorDirection(1)
			case cl.Down:
				driver.SetMotorDirection(-1)
			case cl.Stop:
				driver.SetMotorDirection(0)
				
			}
		case <- slaveStatus:
			break
		}
	}
}


func main(){
	Run()
}
