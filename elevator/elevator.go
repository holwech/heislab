package elevator

import (
	"github.com/holwech/heislab/driver"
)

type ElevData struct {
	InputType string
	InputValue int
}

type ElevatorState struct{
	Floor, Direction, RequestedFloor int
	IsInFloor bool
}

func Run(elevData chan ElevData){
	orderChan := make(chan driver.Order)
	floorChan := make(chan int)
	go driver.ReadOrders(orderChan)
	go driver.ReadFloorSensor(floorChan)

	//Move elevator down until floor is reached
	driver.InitHardware()
	driver.SetMotorDirection(-1)

	for{
		floorVal := <-floorChan
		if floorVal != -1{
			driver.SetMotorDirection(0)
			break
		}
	}

	//Continously read inputs 
	for{
		select{

		}
	}
}

func GoToFloor(floor int) chan bool {
	state.RequestedFloor = floor
}


arrived := GoToFloor(3)
select {
	whgatever
	<-arrived:
		stuff when we have arroved at floor 3
}