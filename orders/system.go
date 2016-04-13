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
	WaitingNextOrder
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

func (sys *System) AddInnerOrder(elevatorIP string, floor int) bool {
	alreadyAdded := false
	elevator, exists := sys.Elevators[elevatorIP]

	if exists {
		if elevator.Orders[floor] != None {
			alreadyAdded = true
		} else {
			elevator.Orders[floor] = Inner
			sys.Elevators[elevatorIP] = elevator
		}
		cmdLight := network.Message{"", elevatorIP, "", cl.LightOnInner, floor}
		sys.Commands <- cmdLight
	}
	return exists && !alreadyAdded
}

func (sys *System) AddOuterOrder(floor, direction int) bool {
	alreadyAdded := false
	if direction == -1 {
		if sys.UnhandledOrdersDown[floor] {
			alreadyAdded = true
		} else {
			sys.UnhandledOrdersDown[floor] = true
		}

		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterDown, floor}
		sys.Commands <- cmdLight
	} else if direction == 1 {
		if sys.UnhandledOrdersUp[floor] {
			alreadyAdded = true
		} else {
			sys.UnhandledOrdersUp[floor] = true
		}
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterUp, floor}
		sys.Commands <- cmdLight
	}
	return !alreadyAdded
}

func (sys *System) RemoveOrder(elevatorIP string, floor int) {
	elevator := sys.Elevators[elevatorIP]
	elevator.Orders[floor] = None
	sys.Elevators[elevatorIP] = elevator
}

func (sys *System) NotifyDoorClosed(elevatorIP string) {
	elevator, exists := sys.Elevators[elevatorIP]
	if !exists {
		return
	}
	sys.SetBehaviour(elevatorIP, Idle)
	for floor := 0; floor < 4; floor++ {
		if elevator.Orders[floor] != None {
			sys.SetBehaviour(elevatorIP, WaitingNextOrder)
		}
	}
}

func (sys *System) SetBehaviour(elevatorIP string, behaviour Behaviour) {
	elevator, exists := sys.Elevators[elevatorIP]
	if !exists {
		return
	}
	elevator.CurrentBehaviour = behaviour
	sys.Elevators[elevatorIP] = elevator
}

func (sys *System) SetDirection(elevatorIP string, direction int) {
	elevator, exists := sys.Elevators[elevatorIP]
	if !exists {
		return
	}
	elevator.Direction = direction
	sys.Elevators[elevatorIP] = elevator
}
