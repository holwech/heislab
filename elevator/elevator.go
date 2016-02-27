package main

import (
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/communication"
	"fmt"
)

func Init() (<-chan driver.InnerOrder,<-chan driver.OuterOrder, <-chan int){
	driver.InitHardware()
	
	innerChan := driver.ListenInnerPanel()
	outerChan := driver.ListenOuterPanel()
	floorChan := driver.ListenFloorSensor()

	//Drive down to first floor
	currentFloor := <-floorChan
	if currentFloor != 1{
		driver.SetMotorDirection(-1)
		for currentFloor != 1{
			currentFloor = <-floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan,outerChan,floorChan
}


// -------- MAIN -- MASTER -- SLAVE? 

//Slave -------------------------------------------
func main(){
	innerChan, outerChan, floorChan := Init()
	driver.SetMotorDirection(1)
	/*
	recvChan := make(chan communication.CommData)
	sendChan := make(chan communication.CommData)
	go communication.Run(recvChan,sendChan)
	*/

//	communication.Send(receiverIP, "COMMAND", "UP", sendChan)
//	comChan := com.Init()

	//Temporary channels for communications between master
	//and slave until communication works
	recv_m := make(chan communication.CommData)
	send_m := make(chan communication.CommData);
	go master(recv_m,send_m)

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
			
		case commData := <-send_m:
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
	}
}



//Master-------------------------
type Behaviour int
const (
	Idle Behaviour = iota
	Moving
	Stopped
)

type ElevatorState struct{
	Floor int
	Direction int
	CurrentBehaviour Behaviour
}

//Listen to inputs from slaves and send actions back
func master(recv chan communication.CommData,send chan communication.CommData){
	/*
	recvChan := make(chan communication.CommData)
	sendChan := make(chan communication.CommData)
	go communication.Run(recvChan,sendChan)
	*/
	
	elevators := make(map[string]ElevatorState)
	testState := ElevatorState{1,0, Idle}
	elevators["localhost"] =testState


	innerOrders := make(map[string][]bool)
	outerOrdersUp := []bool{false,false,false}
	outerOrdersDown := []bool{false,false,false}

	innerOrders["localhost"] = []bool{false,false,false,false}

	for{
		select{
		case commData := <- recv:
			//Decode message, do corresponding action
			switch commData.DataType{
			case "INNER":
				fmt.Printf("Received in master/INNER: ")
				fmt.Println(commData.DataValue)
				order := commData.DataValue.(driver.InnerOrder)
				innerOrders["localhost"][order.Floor-1] = true
			case "OUTER":
				fmt.Printf("Received in master/OUTER: ")
				fmt.Println(commData.DataValue)
				order := commData.DataValue.(driver.OuterOrder)
				if(order.Direction == 1){
					outerOrdersUp[order.Floor-1] = true
				}else{
					outerOrdersDown[order.Floor-2] = true
				}
			case "FLOOR":
				fmt.Printf("Received in master/FLOOR: ")
				fmt.Println(commData.DataValue)
				floor := commData.DataValue.(int)
				var elevator = elevators["localhost"]
				elevator.Floor = floor
				elevators["localhost"] = elevator
				fmt.Println(elevators["localhost"])
				if floor == 1{
					com := communication.CommData{
						DataType:"MOVEUP",
					}
					send <- com
				}else if floor == 4{
					com := communication.CommData{
						DataType:"MOVEDOWN",
					}
					send <- com
				}
			}
		}
	}

}