package orders

import "github.com/holwech/heislab/network"

/*WANT: 


handle orders:
-needs to chose which elevator is going to handle an order
-needs to know which elevator is close and idle or going same direction
-elevator states should be kept in elevator package?
-Do we even need an elevator package? Part of slave?
-pass elevator states to order handling?
-in that case, should be read only


*/

/*
After some thinking-consulting:
orders should be kept in elevator module as inner orders could
be considered part of elevator state
and
outer orders is global
*/


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
	Elevators map[string]ElevatorState
	OuterOrdersUp [4]bool
	OuterOrdersDown [4]bool
}

func NewSystem() *System{
	var s System
	s.Elevators = make(map[string]ElevatorState)
	return &s
}


func (sys *System) AddElevator(elevatorIP string) bool{
	_, exists := sys.Elevators[elevatorIP]
	if !exists{
		sys.Elevators[elevatorIP] = ElevatorState{}
	}
	return !exists
}

func (sys *System) RemoveElevator(elevatorIP string) bool{
	_, exists := sys.Elevators[elevatorIP]
	if exists{
		delete(sys.Elevators[elevatorIP], elevatorIP)
	}
	return exists
}

func (sys *System) AddInnerOrder(elevatorIP string, floor int) bool{
	alreadyAdded := false
	elevator, exists := sys.Elevators[elevatorIP];

	if exists{
		if elevator.InnerOrders[floor]{
			alreadyAdded = true
		}else{
			elevator.InnerOrders[floor] = true	
			sys.Elevators[elevatorIP] = elevator
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

func (sys *System) UpdateFloor(elevatorIP string, floor int){
	elevator := sys.Elevators[elevatorIP]
	elevator.Floor = floor
	sys.Elevators[elevatorIP] = elevator
}


	
func (sys *System) Command() (network.Message){
	var command network.Message;
	
	//Check if elevator is on same floor as an order
	for elev := range sys.Elevators{
		for floor := 0; floor < 4; floor++{
			if elev.InnerOrders[floor] == true && elev.Floor == floor{
				command.Response = cl.Stop
			}
		}
	}

}


