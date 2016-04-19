package scheduler

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

const MAXCOST int = 100000

func (sys *System) NotifyInnerOrder(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]

	if inSystem {
		elevator.InnerOrders[floor] = true
		sys.Elevators[elevatorIP] = elevator
		cmdLight := network.Message{"", elevatorIP, "", cl.LightOnInner, floor}
		sys.Commands <- cmdLight
		if elevator.CurrentBehaviour == Idle{
			sys.SetBehaviour(elevatorIP,AwaitingCommand)
		}
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

func (sys *System) NotifyFloor(elevatorIP string, floor int) {
	elevator, inSystem := sys.Elevators[elevatorIP]
	if inSystem && floor != -1 {
		elevator.Floor = floor
		sys.Elevators[elevatorIP] = elevator
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
				if minCostElev.CurrentBehaviour == Idle{
					sys.SetBehaviour(minCostElevIP,AwaitingCommand)
				}
				minCostElev.OuterOrdersUp[floor] = true
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
				cost := elev.costOfOuterOrder(floor, -1)
				if cost < minCost {
					minCost = cost
					minCostElev = elev
					minCostElevIP = elevIP
				}
			}
			if minCost < MAXCOST {
				if minCostElev.CurrentBehaviour == Idle{
					sys.SetBehaviour(minCostElevIP,AwaitingCommand)
				}
				minCostElev.OuterOrdersDown[floor] = true
				sys.Elevators[minCostElevIP] = minCostElev
				sys.UnhandledOrdersDown[floor] = false
			}
		}
	}
}

func (sys *System) CommandConnectedElevators() {
	for elevIP, elev := range sys.Elevators {
		switch elev.CurrentBehaviour {
		case Idle:
		case Moving:
			if elev.hasOrderAtFloor(elev.Floor){
				sys.sendStopCommands(elevIP)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			}
		case DoorOpen:
			if elev.hasOrderAtFloor(elev.Floor){
				sys.sendStopCommands(elevIP)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			}
		case AwaitingCommand:
			if elev.hasOrderAtFloor(elev.Floor){
				sys.sendStopCommands(elevIP)
				sys.ClearOrder(elevIP, elev.Floor)
				sys.SetBehaviour(elevIP, DoorOpen)
			}else{
				for floor := 0; floor < 4; floor++ {
					var command network.Message
					command.Receiver = elevIP
					if elev.Floor > floor {
						command.Response = cl.Down
						sys.SetDirection(elevIP, -1)
					}else if elev.Floor < floor {
						command.Response = cl.Up
						sys.SetDirection(elevIP, 1)
					}
					sys.SetBehaviour(elevIP, Moving)
					sys.Commands <- command
				}
			}
		}
	}
}


func (elev *ElevatorState) costOfOuterOrder(floor, direction int) int {
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

func (sys *System) sendStopCommands(elevIP string) {
	elevator := sys.Elevators[elevIP]
	var command network.Message

	command.Receiver = elevIP
	command.Response = cl.Stop
	sys.Commands <- command
	if elevator.OuterOrdersDown[elevator.Floor]{
		sys.Commands <- network.Message{"",cl.All,"",cl.LightOffOuterDown,elevator.Floor}
	}
	if elevator.OuterOrdersUp[elevator.Floor]{
		sys.Commands <- network.Message{"",cl.All,"",cl.LightOffOuterUp,elevator.Floor}
	}
	if elevator.InnerOrders[elevator.Floor]{
		sys.Commands <- network.Message{"",elevIP,"",cl.LightOffInner,elevator.Floor}
	}
}

func intAbs(num int) int {
	if num < 0 {
		return -num
	} else {
		return num
	}
}
