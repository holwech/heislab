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
	communication.Init("skal fikses", communicationData)
	elevator.Init(elevatorData)
	for{
		select {
			case elevatorInput := <- elevatorData
				handleElevInput(elevatorInput)
			case dataInput := <- communicationData
				handleDataInput(dataInput)
			case connectionInput := <- connectionData
				handleConnectionInput(connectionInput)
		}
	}
}

func handleElevInput(input elevator.ElevData) {
	fmt.Printf("Input type: %s, input value: %d", input.InputType)
}

func handleDataInput(data communication.CommData) {
	fmt.Printf("Input type: %s, input value: %d", input.InputType)	
}

func handleConnectionInput(connectionData communication.ConnData) {
	fmt.Printf("Input type: %s, input value: %d", input.InputType)
}