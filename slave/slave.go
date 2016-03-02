package main

import (
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/communication"
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

func main(){
	innerChan, outerChan, floorChan := InitElevator()
	//messageChan, statusChan := InitNetwork()
	//sendChan := network get send chan


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