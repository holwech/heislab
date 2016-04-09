package main

import (
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/master"
	"time"
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

func initSlave(slaveSend chan network.Message, slaveReceive chan network.Message) string {
	master = network.LocalIP()
	ticker := time.NewTicker(50 * time.Millisecond).C
	slaveSend <- network.Message{
		Sender: network.LocalIP(),
		Receiver: cl.All,
		ID: CreateID("Slave"),
		Response: cl.Startup,
		Content: time.Now(),
	}
	for{
		flag := false
		select{
		case message := slaveReceive:
			if message.Response == cl.JoinMaster {
				master := message.Sender
				flag = true
				fmt.Println("Master found")
			}
		case <- ticker:
			master := network.LocalIP()
			flag = true
			fmt.Println("No master found")
		}
		if state == cl.Slave {
			break
		}
	}
	slaveSend <- network.Message{
		Sender: network.LocalIP(),
		Receiver: network.LocalIP,
		ID: CreateID("Slave"),
		Response: cl.SetMaster,
		Content: time.Now(),
	}
	return master
}

func Run() {
	innerChan, outerChan, floorChan := InitElevator()
	slaveSend := make(chan network.Message)
	masterSend := make(chan network.Message)
	nw := InitNetwork(slaveSend, masterSend)
	slaveReceive, slaveStatus := nw.SChannels()
	go master.Run(nw, masterSend)
	master := initSlave(slaveSend, slaveReceive)
	var doorTimer *time.Timer

	for {
		select{
		case innerOrder := <- innerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: master
				ID: network.CreateID("Slave"),
				Response: cl.InnerOrder,
				Content: innerOrder,
			}
			slaveSend <- message
		case outerOrder := <- outerChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: master
				ID: network.CreateID("Slave"),
				Response: cl.OuterOrder,
				Content: outerOrder,
			}
			slaveSend <- message
		case newFloor := <- floorChan:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: master
				ID: network.CreateID("Slave"),
				Response: cl.Floor,
				Content: newFloor,
			}
			slaveSend <- message
		case <- doorTimer.C:
			message := network.Message{
				Sender: network.LocalIP(),
				Receiver: master
				ID: network.CreateID("Slave"),
				Response: cl.DoorClosed,
				Content: newFloor,
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
				doorTimer = time.NewTimer(3 * time.Second)
			}
		case <- slaveStatus:
			break
		}
	}
}


func main(){
	Run()
}
