package scheduler

import (
	"fmt"
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
			elevator.AwaitingCommand = true
		}
		sys.Elevators[elevatorIP] = elevator
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
		shouldStop := false
		if elevator.InnerOrders[floor] {
			shouldStop = true
		}
		if elevator.Direction == 1 && elevator.OuterOrdersUp[floor] {
			shouldStop = true
		} else if elevator.Direction == -1 && elevator.OuterOrdersDown[floor] {
			shouldStop = true
		}
		if elevator.Direction == -1 && elevator.OuterOrdersUp[floor] && !hasUpOrdersBelow {
			shouldStop = true
		} else if elevator.Direction == 1 && elevator.OuterOrdersDown[floor] && !hasDownOrdersAbove {
			shouldStop = true
		}
		if shouldStop {
			elevator.AwaitingCommand = false
			sys.Elevators[elevatorIP] = elevator
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
			elevator.AwaitingCommand = true
			fmt.Println("more orders")
		} else {
			sys.SetBehaviour(elevatorIP, Idle)
		}
		sys.Elevators[elevatorIP] = elevator
	}
}

func (sys *System) NotifyEngineFail(elevatorIP string) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem {
		sys.UnassignOuterOrders(elevatorIP)
		if !elevator.hasMoreOrders() {
			if elevator.Direction == 1 {
				elevator.InnerOrders[elevator.Floor+1] = true
			} else {
				elevator.InnerOrders[elevator.Floor-1] = true
			}
			sys.SetBehaviour(elevatorIP, EngineFailure)
		}
		sys.Elevators[elevatorIP] = elevator
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

func (sys *System) CommandConnectedElevators(outgoingCommands chan network.Message) {
	for elevIP, elev := range sys.Elevators {
		if elev.CurrentBehaviour == Idle || elev.CurrentBehaviour == DoorOpen {
			if elev.hasOrderAtFloor(elev.Floor) {
				sys.sendStopCommands(elevIP, outgoingCommands)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			} else {
				if elev.AwaitingCommand {
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
}

func (elev *ElevatorState) costOfOuterOrder(floor, direction int) int {
	if elev.CurrentBehaviour == EngineFailure {
		return MAXCOST
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
