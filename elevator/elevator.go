package main

import (
	"github.com/holwech/heislab/driver"
	"fmt"
)

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

func Init() (<-chan driver.InnerOrder,<-chan driver.OuterOrder, <-chan int){
	driver.InitHardware()
	
	innerChan := driver.ListenInnerPanel()
	outerChan := driver.ListenOuterPanel()
	floorChan := driver.ListenFloorSensor()

	//Drive down to closest floor
	currentFloor := <-floorChan
	if currentFloor == -1{
		driver.SetMotorDirection(-1)
		for currentFloor < 1{
			currentFloor = <-floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan,outerChan,floorChan
}


func main(){
	var eState ElevatorState
	innerChan, outerChan, floorChan := Init()
	
	eState.Floor = <-floorChan
	eState.Direction = 0
	eState.CurrentBehaviour = Idle
	
	for{
		select{
			case innerOrder := <-innerChan:
				fmt.Println(innerOrder)
			case outerOrder := <-outerChan:
				fmt.Println(outerOrder)
			case floorReached := <-floorChan:
				fmt.Println(floorReached)
		}
	}
}
