dpackage main

import (
<<<<<<< HEAD
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/types"
	"fmt"
	"github.com/holwech/heislab/communication"
	"time"
)

func main() {
	request := make(chan int,1)
	request <- 3
	go RunElevator(request)
	neverstop := make(chan int)
	<-neverstop
}

func InitElevator()(innerChan chan types.InnerOrder,outerChan chan types.OuterOrder, floorChan chan int){
	innerChan = make(chan types.InnerOrder)
	outerChan = make(chan types.OuterOrder)
	floorChan = make(chan int)

	driver.InitHardware()
	driver.SetMotorDirection(-1)

	go driver.ReadInnerPanel(innerChan)
	go driver.ReadOuterPanel(outerChan)
	go driver.ReadFloorSensor(floorChan)

	for{
		floorVal := <-floorChan
		if floorVal != -1{
			driver.SetMotorDirection(0)
			return
		}

	}
}

func RunElevator(newRequest chan int){
	innerChan,outerChan, floorChan := InitElevator()
	fmt.Println("Init complete")

	var state types.ElevatorState
	for{
		select{
			case inner := <- innerChan:
				fmt.Println(inner)

			case outer := <- outerChan:
				newRequest <- outer.Floor

			case floor := <-floorChan:
				if floor == -1{
					state.IsInFloor = false
				}else{
					state.Floor = floor
					state.IsInFloor = true
					if state.Floor == state.Request{
						driver.SetMotorDirection(0)
					} 
				}

			case request := <- newRequest:
				state.Request = request
				if state.Floor < state.Request {
					driver.SetMotorDirection(1)
				}else if state.Floor > state.Request{
					driver.SetMotorDirection(-1)
				}
		}
	}
}

func communicationTest() {
	data := map[string]interface{} {
		"LOL": 1,
		"FAKA U BTCH": "U EAT MY NUDLS",
	}
	receiveChannel := make(chan communication.UDPData)
	sendChannel := make(chan communication.UDPData)
	communication.Init("10.20.78.108", receiveChannel, sendChannel)
	time.Sleep(1*time.Second)
	communication.Send("10.20.78.108", data, sendChannel)
	for {
		message := <- receiveChannel
		communication.PrintMessage(&message)
	}
}