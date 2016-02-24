package main

import (
	"github.com/holwech/heislab/driver"
	"fmt"
)

func Init() (<-chan driver.InnerOrder,<-chan driver.OuterOrder, <-chan int){
	driver.InitHardware()
	
	innerChan := driver.ReadInnerPanel()
	outerChan := driver.ReadOuterPanel()
	floorChan := driver.ReadFloorSensor()

	//Drive down to closest floor
	currentFloor := <-floorChan
	if currentFloor == -1{
		driver.SetMotorDirection(-1)
		for currentFloor < 1{
			currentFloor =<- floorChan
		}
		driver.SetMotorDirection(0)
	}

	return innerChan,outerChan,floorChan
}

func main(){
	innerChan, outerChan, floorChan := Init()

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
