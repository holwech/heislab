package scheduler

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

const MAXCOST int = 100000

func (sys *System) NotifyInnerOrder(elevatorIP string, floor int) chan network.Message {
	outgoingCommands := make(chan network.Message, 1)
	defer close(outgoingCommands)
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.InnerOrders[floor] = true
		sys.Elevators[elevatorIP] = elevator
		cmdLight := network.Message{"", elevatorIP, "", cl.LightOnInner, floor}
		outgoingCommands <- cmdLight
		if elevator.CurrentBehaviour == Idle {
			elevator.AwaitingCommand = true
		}
		sys.Elevators[elevatorIP] = elevator
	}
	return outgoingCommands
}

func (sys *System) NotifyOuterOrder(floor, direction int) chan network.Message {
	outgoingCommands := make(chan network.Message, 1)
	defer close(outgoingCommands)
	if direction == -1 {
		sys.UnhandledOrdersDown[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterDown, floor}
		outgoingCommands <- cmdLight
	} else if direction == 1 {
		sys.UnhandledOrdersUp[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterUp, floor}
		outgoingCommands <- cmdLight
	}
	return outgoingCommands
}

func (sys *System) NotifyFloor(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		elevator.Floor = floor

		if elevator.shouldStop(floor) {
			elevator.AwaitingCommand = true
		}
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) NotifyDoorClosed(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		if elevator.HasMoreOrders() {
			elevator.AwaitingCommand = true
			sys.Elevators[elevatorIP] = elevator
		} else {
			sys.SetBehaviour(elevatorIP, Idle)
		}
	}
}

func (sys *System) NotifyEngineFail(elevatorIP string) {
	sys.UnassignOuterOrders(elevatorIP)
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		if !elevator.HasMoreOrders() {
			if elevator.Direction == 1 {
				elevator.InnerOrders[elevator.Floor+1] = true
			} else {
				elevator.InnerOrders[elevator.Floor-1] = true
			}
			sys.Elevators[elevatorIP] = elevator
		}
		sys.Print()
		sys.SetBehaviour(elevatorIP, EngineFailure)
	}
}

func (sys *System) NotifyDisconnectionActive(elevatorIP string) {
	_, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		sys.UnassignOuterOrders(elevatorIP)
		sys.RemoveElevator(elevatorIP)
	}
}

func (sys *System) NotifyDisconnectionInactive(elevatorIP string) {
	_, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		for floor := 0; floor < 4; floor++ {
			sys.UnhandledOrdersDown[floor] = false
			sys.UnhandledOrdersUp[floor] = false
		}
		sys.RemoveElevator(elevatorIP)
	}
}

func (sys *System) AssignOuterOrders() {
	for floor := 0; floor < 4; floor++ {
		if sys.UnhandledOrdersUp[floor] {
			var minCost int = MAXCOST
			var minCostElevIP string
			var minCostElev ElevatorState
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				cost := elev.costOfOuterOrder(floor, 1)
				if (cost < minCost) ||
					(cost == minCost && elevIP < minCostElevIP) {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				minCostElev.OuterOrdersUp[floor] = true
				sys.UnhandledOrdersUp[floor] = false
				if minCostElev.CurrentBehaviour == Idle {
					minCostElev.AwaitingCommand = true
				}
				sys.Elevators[minCostElevIP] = minCostElev
			}
		}
		if sys.UnhandledOrdersDown[floor] {
			var minCost int = MAXCOST
			var minCostElevIP string
			var minCostElev ElevatorState
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				cost := elev.costOfOuterOrder(floor, -1)
				if (cost < minCost) ||
					(cost == minCost && elevIP < minCostElevIP) {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				minCostElev.OuterOrdersDown[floor] = true
				sys.UnhandledOrdersDown[floor] = false
				if minCostElev.CurrentBehaviour == Idle {
					minCostElev.AwaitingCommand = true
				}
				sys.Elevators[minCostElevIP] = minCostElev
			}
		}
	}
}

func (sys *System) CommandConnectedElevators() chan network.Message {
	outgoingCommands := make(chan network.Message, 100)
	defer close(outgoingCommands)
	for elevatorIP, elevator := range sys.Elevators {
		if elevator.AwaitingCommand {
			elevator.AwaitingCommand = false
			sys.Elevators[elevatorIP] = elevator
			if elevator.shouldStop(elevator.Floor) {
				sys.generateStopCommands(elevatorIP, outgoingCommands)
				sys.ClearOrder(elevatorIP, elevator.Floor)
				sys.SetBehaviour(elevatorIP, DoorOpen)
			} else if elevator.CurrentBehaviour == Idle || elevator.CurrentBehaviour == DoorOpen {
				var command network.Message
				command.Receiver = elevatorIP
				if elevator.Direction == 1 {
					command.Response = cl.Down
					sys.SetDirection(elevatorIP, -1)
					for floor := elevator.Floor + 1; floor < 4; floor++ {
						if elevator.hasOrderAtFloor(floor) {
							command.Response = cl.Up
							sys.SetDirection(elevatorIP, 1)
						}
					}
				} else {
					command.Response = cl.Up
					sys.SetDirection(elevatorIP, 1)
					for floor := 0; floor < elevator.Floor; floor++ {
						if elevator.hasOrderAtFloor(floor) {
							command.Response = cl.Down
							sys.SetDirection(elevatorIP, -1)
						}
					}
				}
				outgoingCommands <- command
				sys.SetBehaviour(elevatorIP, Moving)
			}
		}
	}
	return outgoingCommands
}

func (elev *ElevatorState) costOfOuterOrder(floor, direction int) int {
	if elev.CurrentBehaviour == EngineFailure {
		return MAXCOST
	}
	if elev.CurrentBehaviour == Idle {
		return 70 * intAbs(elev.Floor-floor)
	}
	if elev.CurrentBehaviour == Moving {
		if (elev.Floor < floor && elev.Direction == 1) ||
			(elev.Floor > floor && elev.Direction == -1) {
			return 80 * intAbs(elev.Floor-floor)
		}
	}
	if (elev.Floor <= floor && elev.Direction == 1) ||
		(elev.Floor >= floor && elev.Direction == -1) {
		return 100 * intAbs(elev.Floor-floor)
	}
	return MAXCOST
}

func (sys *System) generateStopCommands(elevIP string, outgoingCommands chan network.Message) {
	elevator := sys.Elevators[elevIP]
	var command network.Message

	command.Receiver = elevIP
	command.Response = cl.Stop
	outgoingCommands <- command
	if elevator.OuterOrdersDown[elevator.Floor] {
		outgoingCommands <- network.Message{"", cl.All, "", cl.LightOffOuterDown, elevator.Floor}
	}
	if elevator.OuterOrdersUp[elevator.Floor] {
		outgoingCommands <- network.Message{"", cl.All, "", cl.LightOffOuterUp, elevator.Floor}
	}
	if elevator.InnerOrders[elevator.Floor] {
		outgoingCommands <- network.Message{"", elevIP, "", cl.LightOffInner, elevator.Floor}
	}
}

func intAbs(num int) int {
	if num < 0 {
		return -num
	} else {
		return num
	}
}
