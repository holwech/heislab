package scheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"io/ioutil"
)

type Behaviour int

const (
	Idle Behaviour = iota
	Moving
	DoorOpen
	EngineFailure
)

type ElevatorState struct {
	Floor            int
	Direction        int
	CurrentBehaviour Behaviour
	InnerOrders      [cl.Floors]bool
	OuterOrdersUp    [cl.Floors]bool
	OuterOrdersDown  [cl.Floors]bool
	AwaitingCommand  bool
}

type System struct {
	Elevators           map[string]ElevatorState
	UnhandledOrdersUp   [cl.Floors]bool
	UnhandledOrdersDown [cl.Floors]bool
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
	for floor := 0; floor < cl.Floors; floor++ {
		if elev.hasOrderAtFloor(floor) {
			return true
		}
	}
	return false
}

func (elev *ElevatorState) shouldStop(floor int) bool {
	hasDownOrdersAbove := false
	hasUpOrdersBelow := false
	for f := elev.Floor + 1; f < cl.Floors; f++ {
		if elev.OuterOrdersDown[f] {
			hasDownOrdersAbove = true
		}
	}
	for f := 0; f < elev.Floor; f++ {
		if elev.OuterOrdersUp[f] {
			hasUpOrdersBelow = true
		}
	}
	shouldStop := false
	if elev.InnerOrders[floor] {
		shouldStop = true
	}
	if elev.Direction == 1 && elev.OuterOrdersUp[floor] {
		shouldStop = true
	} else if elev.Direction == -1 && elev.OuterOrdersDown[floor] {
		shouldStop = true
	}
	if elev.Direction == -1 && elev.OuterOrdersUp[floor] && !hasUpOrdersBelow {
		shouldStop = true
	} else if elev.Direction == 1 && elev.OuterOrdersDown[floor] && !hasDownOrdersAbove {
		shouldStop = true
	}
	return shouldStop

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
	for floor := 0; floor < cl.Floors; floor++ {
		s.UnhandledOrdersDown[floor] = sys1.UnhandledOrdersDown[floor] || sys2.UnhandledOrdersDown[floor]
		s.UnhandledOrdersUp[floor] = sys1.UnhandledOrdersUp[floor] || sys2.UnhandledOrdersUp[floor]
	}
	return &s
}

func (sys *System) ToMessage() network.Message {
	backup := network.Message{network.LocalIP(), "",
		network.CreateID(cl.Master), cl.Backup, *sys}
	return backup
}

func SystemFromMessage(message network.Message) *System {
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
						elevTmp.AwaitingCommand = (val2.(bool))
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
		sys.Elevators[elevatorIP] = ElevatorState{Direction: 1}
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
		for floor := 0; floor < cl.Floors; floor++ {
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
		case EngineFailure:
			fmt.Println(" Engine Failure ")
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

func (sys *System) SendLightCommands(outgoingCommands chan network.Message) {
	for floor := 0; floor < cl.Floors; floor++ {
		assignedUp := false
		assignedDown := false
		for elevIP, elev := range sys.Elevators {
			if elev.InnerOrders[floor] {
				outgoingCommands <- network.Message{"", elevIP, "", cl.LightOnInner, floor}
			} else {
				outgoingCommands <- network.Message{"", elevIP, "", cl.LightOffInner, floor}
			}
			if elev.OuterOrdersUp[floor] {
				assignedUp = true
			}
			if elev.OuterOrdersDown[floor] {
				assignedDown = true
			}
		}
		if assignedDown || sys.UnhandledOrdersDown[floor] {
			outgoingCommands <- network.Message{"", cl.All, "", cl.LightOnOuterDown, floor}
		} else {
			outgoingCommands <- network.Message{"", cl.All, "", cl.LightOffOuterDown, floor}
		}
		if assignedUp || sys.UnhandledOrdersUp[floor] {
			outgoingCommands <- network.Message{"", cl.All, "", cl.LightOnOuterUp, floor}
		} else {
			outgoingCommands <- network.Message{"", cl.All, "", cl.LightOffOuterUp, floor}
		}
	}
}

func (sys *System) WriteToFile() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(*sys)
	printError("Encode error: ", err)
	err = ioutil.WriteFile("tmp", buffer.Bytes(), 0644)
	printError("WriteFile error: ", err)
	fmt.Println("Buffer data: ", buffer.String())
}

func ReadFromFile() System {
	var sys System
	file, err := ioutil.ReadFile("tmp")
	printError("ReadFile error: ", err)
	buffer := bytes.NewBuffer(file)
	fmt.Println("Buffer data: ", buffer.String())
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&sys)
	printError("Decode error: ", err)
	return sys
}

func printError(comment string, err error) {
	if err != nil {
		fmt.Println(comment, err)
		fmt.Println(err.Error())
	}
}
