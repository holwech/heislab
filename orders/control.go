package orders

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

func (sys *System) NotifyFloor(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		if elevator.Orders[floor] != None {
			var command network.Message
			var commandLight network.Message

			command.Receiver = elevatorIP
			command.Response = cl.Stop
			sys.Commands <- command

			commandLight.Receiver = elevatorIP
			if elevator.Orders[floor] == OuterDown {
				commandLight.Response = cl.LightOffOuterDown
			} else if elevator.Orders[floor] == OuterUp {
				commandLight.Response = cl.LightOffOuterUp
			}
			commandLight.Content = floor
			sys.Commands <- commandLight
			commandLight.Response = cl.LightOffInner
			sys.Commands <- commandLight

			sys.RemoveOrder(elevatorIP, floor)
			sys.SetBehaviour(elevatorIP, DoorOpen)		
		}
		//Update current floor and stop if order in floor
		elevator.Floor = floor
		sys.Elevators[elevatorIP] = elevator
	}
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
func (sys *System) CheckNewCommand() {
	var command network.Message
	for elevIP := range sys.Elevators {
		elev := sys.Elevators[elevIP]
		if elev.CurrentBehaviour == DoorOpen {
			continue
		}
		for floor := 0; floor < 4; floor++ {
			if elev.Orders[floor] != None &&
				(elev.CurrentBehaviour == Idle || elev.CurrentBehaviour == AwaitingCommand) {
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
					cmdLight := network.Message{"", cl.All, "", cl.LightOffInner, floor}
					sys.Commands <- cmdLight
					cmdLight.Response = cl.LightOffOuterUp
					sys.Commands <- cmdLight
					cmdLight.Response = cl.LightOffOuterDown
					sys.Commands <- cmdLight
				}
				sys.Commands <- command
			}
		}
	}
}
