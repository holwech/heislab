package main

import (
	"github.com/holwech/heislab/elevator"
	"fmt"
	"github.com/holwech/heislab/communication"
)

func main() {
	connectionData := make(chan communication.ConnData)
	communicationData := make(chan communication.CommData)
	elevatorData := make(chan elevator.ElevData)
	communication.Init(communicationData)
	
	elevator.Run(elevatorData)

	elevator.GoToFloor(3)

	for{
		select{
			case elevatorInput := <- elevatorData:
				handleElevInput(elevatorInput)
			case dataInput := <- communicationData:
				handleDataInput(dataInput)
			case connectionInput := <- connectionData:
				handleConnectionInput(connectionInput)
		}
	}
}

func handleElevInput(input elevator.ElevData) {
	fmt.Printf("Input type: %s, input value: %d", input.InputType, input.InputValue)
}

func handleDataInput(input communication.CommData) {
	fmt.Println(input)	
}

func handleConnectionInput(input communication.ConnData) {
	fmt.Println(input)
}