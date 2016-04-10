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
	DoorOpen
)

type Order int
const (
	None Order = iota
	Inner
	OuterUp
	OuterDown
)

type ElevatorState struct{
	Floor int //Previous active floor
	Direction int
	CurrentBehaviour Behaviour
	Orders [4] Order
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

//Deletes an elevator from the system and returns its outer orders to unhandled
func (sys *System) RemoveElevator(elevatorIP string) bool{
	_, exists := sys.Elevators[elevatorIP]
	if exists{
		delete(sys.Elevators, elevatorIP)
	}
	return exists
}

func (sys *System) AddInnerOrder(elevatorIP string, floor int) bool{
	alreadyAdded := false
	elevator, exists := sys.Elevators[elevatorIP];

	if exists{
		if elevator.Orders[floor] != None{
			alreadyAdded = true
		}else{
			elevator.Orders[floor] = Inner	
			sys.Elevators[elevatorIP] = elevator
		}
	}
	return exists && !alreadyAdded
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

func (sys *System) RemoveOrder(elevatorIP string, floor int){
	elevator := sys.Elevators[elevatorIP];
	elevator.Orders[floor] = None
	sys.Elevators[elevatorIP] = elevator
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

	if elevator.Orders[floor] != None{
		sys.RemoveOrder(elevatorIP,floor)
		command.Receiver = elevatorIP
		command.Response = cl.Stop
		sys.SetBehaviour(elevatorIP, DoorOpen)
		return command,true
	}
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

func (sys *System) DoorClosedEvent(elevatorIP string){
	elevator,exists := sys.Elevators[elevatorIP]
	if !exists{
		return
	} 
	sys.SetBehaviour(elevatorIP, Idle)	
	for floor := 0; floor < 4; floor++{
		if elevator.Orders[floor]{
			sys.SetBehaviour(elevatorIP,Moving)
		}
	}
}

func (sys *System) SetBehaviour(elevatorIP string, behaviour Behaviour){
	elevator,exists := sys.Elevators[elevatorIP]
	if !exists{
		return
	}
	elevator.CurrentBehaviour = behaviour
	sys.Elevators[elevatorIP] = elevator
}

//Assign unhandled orders to available elevators
func (sys *System) AssignOrders(){
	for floor := 0; floor < 4; floor++{
		if sys.UnhandledOrdersUp[floor]{
			for elevIP := range sys.Elevators{
				elev := sys.Elevators[elevIP]
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if elev.CurrentBehaviour == Idle || 
						(elev.CurrentBehaviour == Moving && 
						elev.Direction == 1 &&
						elev.Floor < floor){
					elev.Orders[floor] = OuterUp
					sys.Elevators[elevatorIP] = elev
					sys.UnhandledOrdersUp[floor] = false
					break
				}
			}
		}
		if sys.UnhandledOrdersDown[floor]{
			for elevIP := range sys.Elevators{
				elev := sys.Elevators[elevIP]
				//Test if this covers stopped elevators that are supposed to go up afterwards
				if elev.CurrentBehaviour == Idle || 
						(elev.CurrentBehaviour == Moving && 
						elev.Direction == -1 &&
						elev.Floor > floor){
					elev.Orders[floor] = OuterDown
					sys.Elevators[elevatorIP] = elev
					sys.UnhandledOrdersDown[floor] = false
					break
				}
			}
		}
	}
}

/*Send elevators to their assigned orders
Assumes that an elevator has only been assigned orders 
That are on its current path, e.g. if an elevator is moving up
it has no orders below it */
func (sys *System) CommandElevators() (network.Message,bool){
	var command network.Message;
	for elevIP := range sys.Elevators{
		elev := sys.Elevators[elevIP]
		if elev.CurrentBehaviour == DoorOpen{
			continue
		}
		for floor := 0; floor < 4; floor++{
			if elev.Orders[floor]{
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


