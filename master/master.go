package main

import (
	//"github.com/holwech/heislab/network"
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
	InnerOrders [4] bool 
}

	
func getCommand(inner map[string]*ElevatorState, up []bool, down []bool) (network.Message,bool){
	var newCommand bool
	var command network.Message;

	//TODO: Figure out how to generate orders to slave
}


//Listen to inputs from slaves and send actions back
func main(){
	receiveMessage, receiveStatus := network.InitNetwork()
	
	//mutex map to prevent simultaneous RW?
	//Will the elevator states be garbage collected if created and
	// added in a local scope?
	elevatorStates := make(map[string]*ElevatorState)
	outerOrdersUp := []bool{false,false,false,false}
	outerOrdersDown := []bool{false,false,false,false}
	
	//When do we send new orders to elevators?
	for{
	select{
	case message := <- messageChan:
		switch message.Response{
		case "INNER":
			floor := int(message.Content)
			elevator := elevatorStates[message.Sender]
			if elevator.Floor != floor{
				elevator.InnerOrders[floor] = true
			} 
		case "OUTER":
			floor := int(message.Content[0])
			direction := int(message.Content[1])
			if direction == 1{
				outerOrdersUp[floor] = true
			}else if direction == -1{
				outerOrdersDown[floor] = true
			}
		case "FLOOR":
			floor := int(response)
			elevator := elevatorStates[message.Sender]
			elevator.Floor = floor
		}
		command,commandOk := getCommand(inner map[string]*ElevatorState, up []bool, down []bool)
		if commandOk{
			network.Send(command)
		}
	case connStatus := <- statusChan:
		switch connStatus.Response{
		case "NEW":
			//Add elevator
		case "TIMEOUT":
			//Remove elevator from connected
		}
	}
	}
}

/*
func main(){
	elevatorStates := make(map[string]*ElevatorState)

	if (1 == 1){
		
	elevator := &ElevatorState{Floor : 1}
	elevatorStates["localhost"] = elevator	
	}
	fmt.Println(*elevatorStates["localhost"])
	elevatorStates["localhost"].Floor = 2
	fmt.Println(*elevatorStates["localhost"])	
}*/
