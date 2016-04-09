package orders

import (
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
)	
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
	Floor int //Previous active floor
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

func (sys *System) CreateBackup() network.Message{
	backup := network.Message{network.LocalIP(),"",
		network.CreateID(cl.Master),cl.Backup,sys}
	return backup
}

func SystemFromBackup(msg network.Message) *System{
	var s System
	s = msg.Content.(System)
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
		delete(sys.Elevators, elevatorIP)
	}
	return exists
}

func (sys *System) AssignOrder(elevatorIP string, floor int) bool{
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

func (sys *System) RemoveOrder(elevatorIP string, floor int){
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

func (sys *System) FloorAction(elevatorIP string, floor int) (network.Message,bool){
	var command network.Message;

	elevator,exists := sys.Elevators[elevatorIP]
	if !exists || floor == -1{
		return command,false
	}
	//Update current floor and stop if order in floor
	elevator.Floor = floor
	sys.Elevators[elevatorIP] = elevator

	if elevator.Orders[floor]{
		sys.RemoveOrder(elevatorIP,floor)
			command.Receiver = elevatorIP
			command.Response = cl.Stop
			sys.SetBehaviour(elevatorIP, Idle)
			for floor := 0; floor < 4; floor++{
				if sys.Elevators[elevatorIP].Orders[floor]{
					sys.SetBehaviour(elevatorIP,Moving)
				}
			}
			return command,true
	}

	//First stop the elevators that have reached their ordered floor - one at a time
	//Check if an elevator has more orders, and set state accordingly -- tbd
	return command,false

}

func (sys *System) SetDirection(elevatorIP string, direction int){
	elevator,exists := sys.Elevators[elevatorIP]
	if !exists{
		return
	}
	elevator.Direction = direction
	sys.Elevators[elevatorIP] = elevator
}

func (sys *System) SetBehaviour(elevatorIP string, behaviour Behaviour){
	elevator,exists := sys.Elevators[elevatorIP]
	if !exists{
		return
	}
	elevator.CurrentBehaviour = behaviour
	sys.Elevators[elevatorIP] = elevator
}

//Remove orders when they are finished
//Prevent multiple elevators from going for the same order. 
//Let each elevator have a list of orders to take?

func (sys *System) Command() (network.Message,bool){
	var command network.Message;

	//Dispatch unhandled orders to the connected elevators - all?
	for floor := 0; floor < 4; floor++{
		if sys.UnhandledOrdersUp[floor]{
			for elevIP := range sys.Elevators{
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if sys.Elevators[elevIP].CurrentBehaviour == Idle || 
						(sys.Elevators[elevIP].CurrentBehaviour == Moving && 
						sys.Elevators[elevIP].Direction == 1 &&
						sys.Elevators[elevIP].Floor < floor){
					sys.AssignOrder(elevIP,floor)
					sys.UnhandledOrdersUp[floor] = false
					break
				}
			}
		}
		if sys.UnhandledOrdersDown[floor]{
			for elevIP := range sys.Elevators{
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if sys.Elevators[elevIP].CurrentBehaviour == Idle || 
						(sys.Elevators[elevIP].CurrentBehaviour == Moving && 
						sys.Elevators[elevIP].Direction == -1 &&
						sys.Elevators[elevIP].Floor > floor){
					sys.AssignOrder(elevIP,floor)
					sys.UnhandledOrdersDown[floor] = false
					break
				}
			}
		}
	}
	//fmt.Println(sys)
	//Send elevators to their orders
	//Find a way to make sure to not send elevators that have just stopped with door open. Maybe include a state for door open
	//and let the elevators report when they are ready?
	for elevIP := range sys.Elevators{ // if elev.behaviour != door open?
		for floor := 0; floor < 4; floor++{
			if sys.Elevators[elevIP].Orders[floor]{
				command.Receiver = elevIP
				if floor < sys.Elevators[elevIP].Floor{
					command.Response = cl.Down 
					sys.SetDirection(elevIP, -1)
				}else{
					command.Response = cl.Up
					sys.SetDirection(elevIP, 1)
				}
				sys.SetBehaviour(elevIP, Moving)
				return command,true
			}
		}
	}
	return command,false
}


