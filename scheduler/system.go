package scheduler

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen
	AwaitingCommand
)

type ElevatorState struct {
	Floor            int
	Direction        int
	CurrentBehaviour Behaviour
	InnerOrders      [4]bool
	OuterOrdersUp    [4]bool
	OuterOrdersDown  [4]bool
	EngineFail       bool
}

type System struct {
	Elevators           map[string]ElevatorState
	UnhandledOrdersUp   [4]bool
	UnhandledOrdersDown [4]bool
}

func (elev *ElevatorState) hasOrderAtFloor(floor int) bool {
	if elev.InnerOrders[floor] ||
		elev.OuterOrdersDown[floor] ||
		elev.OuterOrdersUp[floor] {
		return true
	}
	return false
}

func (elev *ElevatorState) hasMoreOrders() bool {
	for floor := 0; floor < 4; floor++ {
		if elev.hasOrderAtFloor(floor) {
			return true
		}
	}
	return false
}

func NewSystem() *System {
	var s System
	s.Elevators = make(map[string]ElevatorState)
	return &s
}

func MergeSystems(sys1 *System, sys2 *System) *System {
	var s System
	s.Elevators = make(map[string]ElevatorState)
	for elevIP, elev := range sys1.Elevators {
		s.Elevators[elevIP] = elev
	}
	for elevIP, elev := range sys2.Elevators {
		s.Elevators[elevIP] = elev
	}
	for floor := 0; floor < 4; floor++ {
		s.UnhandledOrdersDown[floor] = sys1.UnhandledOrdersDown[floor] || sys2.UnhandledOrdersDown[floor]
		s.UnhandledOrdersUp[floor] = sys1.UnhandledOrdersUp[floor] || sys2.UnhandledOrdersUp[floor]
	}
	return &s
}

func (sys *System) CreateBackup() network.Message {
	backup := network.Message{network.LocalIP(), "",
		network.CreateID(cl.Master), cl.Backup, *sys}
	return backup
}

func SystemFromBackup(message network.Message) *System {
	s := NewSystem()
	for key, val := range message.Content.(map[string]interface{}) {
		switch val.(type) {
		case map[string]interface{}:
			for elevIP, elevInterface := range val.(map[string]interface{}) {
				var elevTmp ElevatorState
				fmt.Println(elevIP, elevInterface)
				for key2, val2 := range elevInterface.(map[string]interface{}) {
					switch val2.(type) {
					case float64:
						switch key2 {
						case "Direction":
							elevTmp.Direction = int(val2.(float64))
						case "Floor":
							elevTmp.Floor = int(val2.(float64))
						case "Behaviour":
							elevTmp.CurrentBehaviour = Behaviour(val2.(float64))
						}
					case []interface{}:
						for i, order := range val2.([]interface{}) {
							if key == "OuterOrdersUp" {
								elevTmp.OuterOrdersUp[i] = order.(bool)
							} else if key == "OuterOrdersDown" {
								elevTmp.OuterOrdersDown[i] = order.(bool)
							} else {
								elevTmp.InnerOrders[i] = order.(bool)
							}
						}
					case bool:
						elevTmp.EngineFail = (val2.(bool))
					}
				}
				s.Elevators[elevIP] = elevTmp
			}
		case []interface{}:
			for i, order := range val.([]interface{}) {
				if key == "UnhandledOrdersUp" {
					s.UnhandledOrdersUp[i] = order.(bool)
				} else if key == "UnhandledOrdersDown" {
					s.UnhandledOrdersDown[i] = order.(bool)
				}

			}
		}
	}
	return s
}

func (sys *System) AddElevator(elevatorIP string) bool {
	_, exists := sys.Elevators[elevatorIP]
	if !exists {
		sys.Elevators[elevatorIP] = ElevatorState{}
	}
	return !exists
}

func (sys *System) RemoveElevator(elevatorIP string) bool {
	_, exists := sys.Elevators[elevatorIP]
	if exists {
		delete(sys.Elevators, elevatorIP)
	}
	return exists
}

func (sys *System) ClearOrder(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.InnerOrders[floor] = false
		elevator.OuterOrdersDown[floor] = false
		elevator.OuterOrdersUp[floor] = false
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) UnassignOuterOrders(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		for floor := 0; floor < 4; floor++ {
			if elevator.OuterOrdersUp[floor] {
				sys.UnhandledOrdersUp[floor] = true
				elevator.OuterOrdersUp[floor] = false
			}
			if elevator.OuterOrdersDown[floor] {
				sys.UnhandledOrdersDown[floor] = true
				elevator.OuterOrdersDown[floor] = false
			}
		}
		sys.Elevators[elevatorIP] = elevator
	}
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

func (sys *System) Print() {
	fmt.Println()
	for elevatorIP, elevator := range sys.Elevators {
		fmt.Printf("%s: floor: %d, direction: %d,", elevatorIP, elevator.Floor, elevator.Direction)
		switch elevator.CurrentBehaviour {
		case Idle:
			fmt.Println(" Idle ")
		case Moving:
			fmt.Println(" Moving ")
		case DoorOpen:
			fmt.Println(" DoorOpen ")
		case AwaitingCommand:
			fmt.Println(" AwaitingCommand ")
		}
		fmt.Print("Inner orders: ", elevator.InnerOrders)
		fmt.Print("  Outer up: ", elevator.OuterOrdersUp)
		fmt.Println("  Outer Down: ", elevator.OuterOrdersDown)
		fmt.Println("--------------------------")
	}
	fmt.Println(sys.UnhandledOrdersUp)
	fmt.Println(sys.UnhandledOrdersDown)
	fmt.Println("--------------------------")

}
