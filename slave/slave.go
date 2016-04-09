package main

import (
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/master"
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
		for currentFloor != 1{
			currentFloor = <-floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan,outerChan,floorChan
}

func InitNetwork(slaveSend chan <- Message, masterSend chan <- Message) *network.Network {
	nw := new(network.Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)
	return nw
}

func Run() {
	innerChan, outerChan, floorChan := InitElevator()
	slaveSend = make(chan network.Message)
	masterSend = make(chan network.Message)
	nw := InitNetwork(slaveSend, masterSend)
	slaveReceive, slaveStatus := nw.SChannels()
	master.Run(nw, masterSend)

	for {
		select{
		case innerOrder := <- innerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: network.LocalIP(),
				ID: network.CreateID("Slave"),
				Response: cl.InnerOrder,
				Content: innerOrder,
			}
			slaveSend <- message
		case outerOrder := <- outerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: network.LocalIP(),
				ID: network.CreateID("Slave"),
				Response: cl.OuterOrder,
				Content: outerOrder,
			}
			slaveSend <- message
		case newFloor := <- floorChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: network.LocalIP(),
				ID: network.CreateID("Slave"),
				Response: cl.Floor,
				Content: newFloor,
			}
			slaveSend <- message
		case message := <- slaveReceive:
			select message.Response{
			case "MOVEUP":
				driver.SetMotorDirection(1)
			case "MOVEDOWN":
				driver.SetMotorDirection(-1)
			case "STOP":
				driver.SetMotorDirectioin(0)
			}
		case status := <- slaveStatus:
			break
	}
}

func main(){
	innerChan, outerChan, floorChan := InitElevator()


	for{
		select{
		case innerOrder := <-innerChan:
			com := communication.CommData{
				DataType:"INNER",
				DataValue:innerOrder,
			}
			recv_m <- com

		case outerOrder := <-outerChan:
			com := communication.CommData{
				DataType:"OUTER",
				DataValue:outerOrder,
			}
			recv_m <- com
			
		case newFloor := <-floorChan:
			com := communication.CommData{
				DataType:"FLOOR",
				DataValue:newFloor,
			}
			recv_m <- com
			
		case commData := <-messageChan:
			fmt.Println(commData)
			switch commData.DataType{
			case "MOVEUP":
				driver.SetMotorDirection(1)
			case "MOVEDOWN":
				driver.SetMotorDirection(-1)
			case "STOP":
				driver.SetMotorDirection(0)
			}
		}
		case connStatus := <-statusChan:
			if no master:
				spawn master
	}
}
