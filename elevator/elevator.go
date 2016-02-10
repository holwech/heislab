package elevator

import (
	"fmt"
	"github.com/holwech/heislab/types"
	"github.com/holwech/heislab/driver"
)

type ElevData struct {
	InputType string
	InputValue int
}

func Init(a chan ElevData){
	//Penis
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