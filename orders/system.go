package orders

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
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
	AwaitingCommand
)

type Order int

const (
	None Order = iota
	Inner
	OuterUp
	OuterDown
)

type ElevatorState struct {
	Floor            int //Previous active floor
	Direction        int
	CurrentBehaviour Behaviour
	Orders           [4]Order
}

type System struct {
	Elevators           map[string]ElevatorState
	UnhandledOrdersUp   [4]bool
	UnhandledOrdersDown [4]bool
	Commands            chan network.Message
}

func NewSystem() *System {
	var s System
	s.Commands = make(chan network.Message, 10)
	s.Elevators = make(map[string]ElevatorState)
	return &s
}

func (sys *System) CreateBackup() network.Message {
	backup := network.Message{network.LocalIP(), "",
		network.CreateID(cl.Master), cl.Backup, sys}
	return backup
}

func SystemFromBackup(msg network.Message) *System {
	var s System
	s = msg.Content.(System)
	return &s
}

func (sys *System) AddElevator(elevatorIP string) bool {
	_, exists := sys.Elevators[elevatorIP]
	if !exists {
		sys.Elevators[elevatorIP] = ElevatorState{}
	}
	return !exists
}

//Deletes an elevator from the system and returns its outer orders to unhandled
func (sys *System) RemoveElevator(elevatorIP string) bool {
	_, exists := sys.Elevators[elevatorIP]
	if exists {
		delete(sys.Elevators, elevatorIP)
	}
	return exists
}

func (sys *System) NotifyInnerOrder(elevatorIP string, floor int){
	elevator, inSystem := sys.Elevators[elevatorIP]

	if inSystem {
		if elevator.Orders[floor] == None {
			elevator.Orders[floor] = Inner
			sys.Elevators[elevatorIP] = elevator
		}
		//if elevator.Floor != floor{
			cmdLight := network.Message{"", elevatorIP, "", cl.LightOnInner, floor}
			sys.Commands <- cmdLight
		//}
	}
}

func (sys *System) NotifyOuterOrder(floor, direction int) {
	if direction == -1 {
		sys.UnhandledOrdersDown[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterDown, floor}
		sys.Commands <- cmdLight
	} else if direction == 1 {
		sys.UnhandledOrdersUp[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterUp, floor}
		sys.Commands <- cmdLight
	}
}

func (sys *System) RemoveOrder(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem{
		elevator.Orders[floor] = None
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) NotifyDoorClosed(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		if elevator.hasMoreOrders(){
			sys.SetBehaviour(elevatorIP,AwaitingCommand)
		}else{
			sys.SetBehaviour(elevatorIP,Idle)
		}
	}
}

func (elev *ElevatorState) hasMoreOrders() bool{
	for floor := 0; floor < 4; floor++ {
		if elev.Orders[floor] != None {
			return true
		}
	}
	return false
}

func (sys *System) SetBehaviour(elevatorIP string, behaviour Behaviour) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.CurrentBehaviour = behaviour
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) SetDirection(elevatorIP string, direction int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.Direction = direction
		sys.Elevators[elevatorIP] = elevator
	}
}
