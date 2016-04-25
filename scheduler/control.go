package scheduler

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

const MAXCOST int = 100000

func (sys *System) NotifyInnerOrder(elevatorIP string, floor int, outgoingCommands chan network.Message) {
	elevator, inSystem := sys.Elevators[elevatorIP]

	if inSystem {
		elevator.InnerOrders[floor] = true
		sys.Elevators[elevatorIP] = elevator
		cmdLight := network.Message{"", elevatorIP, "", cl.LightOnInner, floor}
		outgoingCommands <- cmdLight
		if elevator.CurrentBehaviour == Idle {
			sys.SetBehaviour(elevatorIP, AwaitingCommand)
		}
	}
}

func (sys *System) NotifyOuterOrder(floor, direction int, outgoingCommands chan network.Message) {
	if direction == -1 {
		sys.UnhandledOrdersDown[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterDown, floor}
		outgoingCommands <- cmdLight
	} else if direction == 1 {
		sys.UnhandledOrdersUp[floor] = true
		cmdLight := network.Message{"", cl.All, "", cl.LightOnOuterUp, floor}
		outgoingCommands <- cmdLight
	}
}

func (sys *System) NotifyFloor(elevatorIP string, floor int, outgoingCommands chan network.Message) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		elevator.Floor = floor
		sys.Elevators[elevatorIP] = elevator

		hasDownOrdersAbove := false
		hasUpOrdersBelow := false
		for f := elevator.Floor + 1; f < 4; f++ {
			if elevator.OuterOrdersDown[f] {
				hasDownOrdersAbove = true
			}
		}
		for f := 0; f < elevator.Floor; f++ {
			if elevator.OuterOrdersUp[f] {
				hasUpOrdersBelow = true
			}
		}
		shouldStop := elevator.hasOrderAtFloor(floor)
		if elevator.Direction == 1 && hasDownOrdersAbove {
			shouldStop = false
		} else if elevator.Direction == -1 && hasUpOrdersBelow {
			shouldStop = false
		}
		if shouldStop {
			sys.sendStopCommands(elevatorIP, outgoingCommands)
			sys.ClearOrder(elevatorIP, floor)
			sys.SetBehaviour(elevatorIP, DoorOpen)
		}
	}
}

func (sys *System) NotifyDoorClosed(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		if elevator.hasMoreOrders() {
			sys.SetBehaviour(elevatorIP, AwaitingCommand)
		} else {
			sys.SetBehaviour(elevatorIP, Idle)
		}
	}
}

func (sys *System) NotifyEngineFail(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.EngineFail = true
		sys.UnassignOuterOrders(elevatorIP)
		if !elevator.hasMoreOrders() {
			if elevator.Direction == 1 {
				elevator.InnerOrders[elevator.Floor+1] = true
			} else {
				elevator.InnerOrders[elevator.Floor-1] = true
			}
			sys.SetBehaviour(elevatorIP, AwaitingCommand)
		}
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) NotifyEngineOk(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		elevator.EngineFail = false
		sys.Elevators[elevatorIP] = elevator
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
				if cost < minCost {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				if minCostElev.CurrentBehaviour == Idle {
					sys.SetBehaviour(minCostElevIP, AwaitingCommand)
				}
				minCostElev.OuterOrdersUp[floor] = true
				sys.Elevators[minCostElevIP] = minCostElev
				sys.UnhandledOrdersUp[floor] = false
				if minCostElev.CurrentBehaviour == Idle {
					sys.SetBehaviour(minCostElevIP, AwaitingCommand)
				}
			}
		}
		if sys.UnhandledOrdersDown[floor] {
			var minCost int = MAXCOST
			var minCostElevIP string
			var minCostElev ElevatorState
			for elevIP := range sys.Elevators {
				elev := sys.Elevators[elevIP]
				cost := elev.costOfOuterOrder(floor, -1)
				if cost < minCost {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				if minCostElev.CurrentBehaviour == Idle {
					sys.SetBehaviour(minCostElevIP, AwaitingCommand)
				}
				minCostElev.OuterOrdersDown[floor] = true
				sys.Elevators[minCostElevIP] = minCostElev
				sys.UnhandledOrdersDown[floor] = false
				if minCostElev.CurrentBehaviour == Idle {
					sys.SetBehaviour(minCostElevIP, AwaitingCommand)
				}
			}
		}
	}
}

func (sys *System) CommandConnectedElevators(outgoingCommands chan network.Message) {
	for elevIP, elev := range sys.Elevators {
		switch elev.CurrentBehaviour {
		case Idle:
		case Moving:
		case DoorOpen:
			if elev.hasOrderAtFloor(elev.Floor) {
				sys.sendStopCommands(elevIP, outgoingCommands)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			}
		case AwaitingCommand:
			if elev.hasOrderAtFloor(elev.Floor) {
				sys.sendStopCommands(elevIP, outgoingCommands)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			} else {
				var command network.Message
				command.Receiver = elevIP
				if elev.Direction == 1 {
					command.Response = cl.Down
					sys.SetDirection(elevIP, -1)
					for floor := elev.Floor + 1; floor < 4; floor++ {
						if elev.hasOrderAtFloor(floor) {
							command.Response = cl.Up
							sys.SetDirection(elevIP, 1)
						}
					}
				} else {
					command.Response = cl.Up
					sys.SetDirection(elevIP, 1)
					for floor := 0; floor < elev.Floor; floor++ {
						if elev.hasOrderAtFloor(floor) {
							command.Response = cl.Down
							sys.SetDirection(elevIP, -1)
						}
					}
				}
				outgoingCommands <- command
				sys.SetBehaviour(elevIP, Moving)
			}
		}
	}
}

func (elev *ElevatorState) costOfOuterOrder(floor, direction int) int {
	if elev.EngineFail {
		return MAXCOST
	}
	hasDownOrdersAbove := false
	hasUpOrdersBelow := false
	for floor := elev.Floor + 1; floor < 4; floor++ {
		if elev.OuterOrdersDown[floor] {
			hasDownOrdersAbove = true
		}
	}
	for floor := 0; floor < elev.Floor; floor++ {
		if elev.OuterOrdersUp[floor] {
			hasUpOrdersBelow = true
		}
	}

	switch elev.CurrentBehaviour {
	case Idle:
		return 100 * intAbs(elev.Floor-floor)
	case Moving:
		if elev.Direction == direction &&
			((elev.Direction == 1 && elev.Floor < floor) ||
				(elev.Direction == -1 && elev.Floor > floor)) {
			return 10 * intAbs(elev.Floor-floor)
		} else if elev.Floor < floor && elev.Direction == 1 && hasDownOrdersAbove {
			return 9 * intAbs(elev.Floor-floor)
		} else if elev.Floor > floor && elev.Direction == -1 && hasUpOrdersBelow {
			return 9 * intAbs(elev.Floor-floor)
		} else {
			return MAXCOST
		}
	case DoorOpen:
		if elev.hasMoreOrders() {
			if elev.Direction == direction &&
				((elev.Direction == 1 && elev.Floor < floor) ||
					(elev.Direction == -1 && elev.Floor > floor)) {
				return 15 * intAbs(elev.Floor-floor)
			} else if elev.Floor < floor && elev.Direction == 1 && hasDownOrdersAbove {
				return 9 * intAbs(elev.Floor-floor)
			} else if elev.Floor > floor && elev.Direction == -1 && hasUpOrdersBelow {
				return 9 * intAbs(elev.Floor-floor)
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
		} else if elev.Floor < floor && elev.Direction == 1 && hasDownOrdersAbove {
			return 9 * intAbs(elev.Floor-floor)
		} else if elev.Floor > floor && elev.Direction == -1 && hasUpOrdersBelow {
			return 9 * intAbs(elev.Floor-floor)
		} else {
			return MAXCOST
		}
	default:
		return MAXCOST
	}
}

func (sys *System) sendStopCommands(elevIP string, outgoingCommands chan network.Message) {
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
