package main

import (
	"github.com/holwech/heislab/driver"
)


/*
TODO:
Make handy datatypes, like order,direction, numFloors etc
*/

func handleFloorIndicator(){
	floorChan := make(chan int)
	go driver.ReadFloorSignal(floorChan)

	for{
		floor := <-floorChan
		if floor != -1{
			driver.SetFloorIndicatorLamp(floor)
		}
	}
}

func main(){
	driver.InitHardware()
	outerBtnChan := make(chan int,1)
	directionChan := make(chan int, 1)
	floorChan := make(chan int,1)

	go driver.ReadFloorSignal(floorChan)
	go driver.ReadOuterPanel(outerBtnChan,directionChan)
	go handleFloorIndicator()

	for{
		floor := <-floorChan
		if floor == 3{
			driver.SetMotorDirection(-1)
		}else if floor == 0{
			driver.SetMotorDirection(1)
		}
	}
}