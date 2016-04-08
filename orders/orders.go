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
	Orders [4] bool
}

type System struct{
	Elevators map[string]ElevatorState
	UnhandledOrdersUp [4]bool
	UnhandledOrdersDown [4]bool
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

//Should we move an elevators orders back to unhandled?
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
		if elevator.Orders[floor]{
			alreadyAdded = true
		}else{
			elevator.Orders[floor] = true	
			sys.Elevators[elevatorIP] = elevator
		}
	}
	return exists && !alreadyAdded
}

func (sys *System) RemoveInnerOrder(elevatorIP string, floor int){
	elevator := sys.Elevators[elevatorIP];
	elevator.Orders[floor] = false
	sys.Elevators[elevatorIP] = elevator
}

func (sys *System) AddOuterOrder(floor, direction int) bool{
	alreadyAdded := false
	if direction == -1{
		if sys.UnhandledOrdersDown[floor]{
			alreadyAdded = true
		}else{
			sys.UnhandledOrdersDown[floor] = true
		}
	}else if direction == 1{
		if sys.UnhandledOrdersUp[floor]{
			alreadyAdded = true
		}else{
			sys.UnhandledOrdersUp[floor] = true
		}
	}
	return !alreadyAdded
}

func (sys *System) UpdateFloor(elevatorIP string, floor int){
	elevator := sys.Elevators[elevatorIP]
	elevator.Floor = floor
	sys.Elevators[elevatorIP] = elevator
}


//Remove orders when they are finished
//Prevent multiple elevators from going for the same order. 
//Let each elevator have a list of orders to take?

func (sys *System) Command() (network.Message){
	var command network.Message;
	

	//Remove finished orders
	for floor := 0; floor < 4; floor++{
		for elev := range sys.Elevators{
			if elev.Orders[floor] == true && elev.Floor == floor{
				sys.RemoveInnerOrder(elev,floor)
				command.Receiver = elev
				command.Response = cl.Stop
				return
			}
		} 
	}


	//Handle current orders
	for 

}



func (sys *System) Command() (network.Message,bool){
	var command network.Message;
	//First stop the elevators that have reached their ordered floor - one at a time
	//Check if an elevator has more orders, and set state accordingly -- tbd
	for elev := range sys.Elevators{
		if elev.Orders[elev.Floor] == true{
			sys.RemoveInnerOrder(elev,elev.Floor)
			command.Receiver = elev	
			command.Response = cl.Stop
			return command,true
		}
	}
	//Then dispatch unhandled orders to the connected elevators - all?
	for floor := 0; floor < 4; floor++{
		if sys.UnhandledOrdersUp[floor]{
			for elev := range sys.Elevators{
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if elev.Behaviour == Idle || elev.Direction == 1{
					sys.AddInnerOrder(elev,floor)
					sys.UnhandledOrdersUp[floor] = false
				}
			}
		}
		if sys.UnhandledOrdersDown[floor]{
			for elev := range sys.Elevators{
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if elev.Behaviour == Idle || elev.Direction == -1{
					sys.AddInnerOrder(elev,floor)
					sys.UnhandledOrdersDown[floor] = false
				}
			}
		}
	}
	//Finally send elevators to their orders
	//Find a way to make sure to not send elevators that have just stopped with door open. Maybe include a state for door open
	//and let the elevators report when they are ready?
	for elev := range sys.Elevators{
		for floor := 0; floor < 4; floor++{
			if elev.Orders[floor]{
				command.Receiver = elev
				command.Response = cl.Move 
				if floor < elev.Floor{
					command.Content = cl.Down
				}
				else{
					command.Content = cl.Up
				}
				return command,true
			}
		}
	}
	return command,false
}


