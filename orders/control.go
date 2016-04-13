package orders

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

func (sys *System) FloorAction(elevatorIP string, floor int) (network.Message, bool) {
	var command network.Message

	elevator, exists := sys.Elevators[elevatorIP]
	if !exists || floor == -1 {
		return command, false
	}
	//Update current floor and stop if order in floor
	elevator.Floor = floor
	sys.Elevators[elevatorIP] = elevator

	if elevator.Orders[floor] != None {
		sys.RemoveOrder(elevatorIP, floor)
		command.Receiver = elevatorIP
		command.Response = cl.Stop
		//Might want to rewrite this, direction is used for light setting in master
		command.Content = elevator.Direction
		sys.SetBehaviour(elevatorIP, DoorOpen)
		return command, true
	}
	return command, false
}

func (sys *System) AssignOrders() {
	for floor := 0; floor < 4; floor++ {
		if sys.UnhandledOrdersUp[floor] {
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				if elev.CurrentBehaviour == Idle ||
					(elev.CurrentBehaviour == Moving &&
						elev.Direction == 1 &&
						elev.Floor < floor) {
					elev.Orders[floor] = OuterUp
					sys.Elevators[elevIP] = elev
					sys.UnhandledOrdersUp[floor] = false
					break
				}
			}
		}
		if sys.UnhandledOrdersDown[floor] {
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				if elev.CurrentBehaviour == Idle ||
					(elev.CurrentBehaviour == Moving &&
						elev.Direction == -1 &&
						elev.Floor > floor) {
					elev.Orders[floor] = OuterDown
					sys.Elevators[elevIP] = elev
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
func (sys *System) GetNextCommand() (network.Message, bool) {
	var command network.Message
	for elevIP := range sys.Elevators {
		elev := sys.Elevators[elevIP]
		if elev.CurrentBehaviour == DoorOpen {
			continue
		}
		for floor := 0; floor < 4; floor++ {
			if elev.Orders[floor] != None {
				command.Receiver = elevIP
				if floor < elev.Floor {
					command.Response = cl.Down
					sys.SetDirection(elevIP, -1)
					sys.SetBehaviour(elevIP, Moving)

				} else if floor > elev.Floor {
					command.Response = cl.Up
					sys.SetDirection(elevIP, 1)
					sys.SetBehaviour(elevIP, Moving)
				} else if floor == elev.Floor && elev.CurrentBehaviour != Moving {
					sys.RemoveOrder(elevIP, floor)
					command.Response = cl.Stop
					sys.SetBehaviour(elevIP, DoorOpen)
				}
				return command, true
			}
		}
	}

	return command, false
}
