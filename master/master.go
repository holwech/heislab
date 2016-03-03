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

type System struct{
	ElevatorStates map[string]ElevatorState
	OuterOrdersUp [4]bool
	OuterOrdersDown [4]bool
}

func NewSystem() *System{
	var s System
	s.ElevatorStates = make(map[string]ElevatorState)
	return &s
}

func (sys *System) AddInnerOrder(elevatorIP string, floor int) bool{
	alreadyAdded := false
	elevator, exists := sys.ElevatorStates[elevatorIP];

	if exists{
		if elevator.InnerOrders[floor]{
			alreadyAdded = true
		}else{
			elevator.InnerOrders[floor] = true	
			sys.ElevatorStates[elevatorIP] = elevator
		}
	}
	return exists && !alreadyAdded
}

func (sys *System) AddOuterOrder(floor, direction int) bool{
	alreadyAdded := false
	if direction == -1{
		if sys.OuterOrdersDown[floor]{
			alreadyAdded = true
		}else{
			sys.OuterOrdersDown[floor] = true
		}
	}else if direction == 1{
		if sys.OuterOrdersUp[floor]{
			alreadyAdded = true
		}else{
			sys.OuterOrdersUp[floor] = true
		}
	}
	return !alreadyAdded
}

func (sys *System) AddElevator(elevatorIP string) bool{
	_, alreadyAdded := sys.ElevatorStates[elevatorIP]
	if !alreadyAdded{
		sys.ElevatorStates[elevatorIP] = ElevatorState{}
	}
	return !alreadyAdded
}

func (sys *System) UpdateFloor(elevatorIP string, floor int){
	elevator := sys.ElevatorStates[elevatorIP]
	elevator.Floor = floor
	sys.ElevatorStates[elevatorIP] = elevator
}


	
func (sys *System) GenerateSlaveCommand() (network.Message){
	var command network.Message;

	//TODO: Figure out how to generate orders to slave
}


//Listen to inputs from slaves and send actions back
func main(){
	receiveMessage, receiveStatus := network.InitNetwork()
	sys := NewSystem()
	//mutex map to prevent simultaneous RW?
	//Will the elevator states be garbage collected if created and
	// added in a local scope?
	
	

	//When do we send new orders to elevators?
	for{
	select{
	case message := <- messageChan:		
		switch message.Response{
		case "INNER":
			sys.AddInnerOrder(message.Sender, int(message.Content))
		case "OUTER":
			ok := sys.AddOuterOrder(int(message.Content)[0],int(message.Content)[1])
		case "FLOOR":
			sys.UpdateFloor(message.Sender,int(message.Content))
		case "MOVEMENT":
			//update elevator state
		}
		command := sys.GetSlaveCommand()
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
