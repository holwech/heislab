package orders

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

const MAXCOST int = 100000

func (sys *System) NotifyFloor(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		if elevator.Orders[floor] != None {
			var command network.Message
			command.Receiver = elevatorIP
			command.Response = cl.Stop
			sys.Commands <- command

			var commandLight network.Message
			commandLight.Receiver = cl.All
			if elevator.Orders[floor] == OuterDown {
				commandLight.Response = cl.LightOffOuterDown
			} else if elevator.Orders[floor] == OuterUp {
				commandLight.Response = cl.LightOffOuterUp
			}
			commandLight.Content = floor
			sys.Commands <- commandLight

			commandLight.Receiver = elevatorIP
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

func intAbs(num int) int {
	if num < 0 {
		return -num
	} else {
		return num
	}
}

func (elev *ElevatorState) CalculateCost(floor, direction int) int {
	switch elev.CurrentBehaviour {
	case Idle:
		return 100 * intAbs(elev.Floor-floor)
	case Moving:
		if elev.Direction == direction &&
			((elev.Direction == 1 && elev.Floor < floor) ||
				(elev.Direction == -1 && elev.Floor > floor)) {
			return 10 * intAbs(elev.Floor-floor)
		} else {
			return MAXCOST
		}
	case DoorOpen:
		if elev.hasMoreOrders() {
			if elev.Direction == direction &&
				((elev.Direction == 1 && elev.Floor < floor) ||
					(elev.Direction == -1 && elev.Floor > floor)) {
				return 15 * intAbs(elev.Floor-floor)
			} else {
				return MAXCOST
			}
		} else {
			return 100 * intAbs(elev.Floor-floor)
		}
	case AwaitingCommand:
		if elev.Direction == direction &&
			((elev.Direction == 1 && elev.Floor < floor) ||
				(elev.Direction == -1 && elev.Floor > floor)) {
			return 12 * intAbs(elev.Floor-floor)
		} else {
			return MAXCOST
		}
	default:
		return MAXCOST
	}
}

func (sys *System) AssignOrders() {
	for floor := 0; floor < 4; floor++ {
		if sys.UnhandledOrdersUp[floor] {
			var minCost int = MAXCOST
			var minCostElevIP string
			var minCostElev ElevatorState
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				cost := elev.CalculateCost(floor, 1)
				if cost < minCost {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				minCostElev.Orders[floor] = OuterUp
				sys.Elevators[minCostElevIP] = minCostElev
				sys.UnhandledOrdersUp[floor] = false
			}
		}
		if sys.UnhandledOrdersDown[floor] {
			var minCost int = MAXCOST
			var minCostElevIP string
			var minCostElev ElevatorState
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				cost := elev.CalculateCost(floor, -1)
				if cost < minCost {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				minCostElev.Orders[floor] = OuterDown
				sys.Elevators[minCostElevIP] = minCostElev
				sys.UnhandledOrdersDown[floor] = false
			}
		}
	}
}

/*Send elevators to their assigned orders
Assumes that an elevator has only been assigned orders
That are on its current path, e.g. if an elevator is moving up
it has no orders below it */
func (sys *System) CheckNewCommand() {
	for elevIP := range sys.Elevators {
		elev := sys.Elevators[elevIP]
		if elev.CurrentBehaviour == DoorOpen {
			if elev.Orders[elev.Floor] != None {
				var command network.Message
				command.Response = cl.Stop
				command.Receiver = elevIP
				sys.RemoveOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
				sys.Commands <- command
				continue
			}
		}
		for floor := 0; floor < 4; floor++ {
			if elev.Orders[floor] != None &&
				(elev.CurrentBehaviour == Idle || elev.CurrentBehaviour == AwaitingCommand) {
				var command network.Message

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

func (sys *System) CommandElevators() {
	for elevIP, elev := range sys.Elevators {
		switch elev.CurrentBehaviour {
		case Idle:
			if elev.Orders[elev.Floor] {
				sys.StopElevator(elevIP)
				sys.RemoveOrder(elevatorIP, elev.Floor)
			}
		case Moving:
		case DoorOpen:
		case AwaitingCommand:
		}
	}
}

func (sys *System) NotifyFloor2(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		elevator.Floor = floor
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) StopElevator(elevIP string) {
	elevator = sys.Elevators[elevIP]
	var command network.Message
	command.Receiver = elevatorIP
	command.Response = cl.Stop
	sys.Commands <- command

	var commandLight network.Message
	commandLight.Receiver = cl.All
	if elevator.Orders[floor] == OuterDown {
		commandLight.Response = cl.LightOffOuterDown
	} else if elevator.Orders[floor] == OuterUp {
		commandLight.Response = cl.LightOffOuterUp
	}
	commandLight.Content = floor
	sys.Commands <- commandLight

	commandLight.Receiver = elevatorIP
	commandLight.Response = cl.LightOffInner
	sys.Commands <- commandLight

	sys.SetBehaviour(elevatorIP, DoorOpen)
}
