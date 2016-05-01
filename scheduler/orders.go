package scheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"io/ioutil"
)

func (sys *System) ClearOuterOrders() []network.Message {
	var orders []network.Message
	for i, _ := range sys.OrdersUp {
		if sys.OrdersUp[i] {
			sys.OrdersUp[i] = false
			message := network.Message{ "", cl.All, "", cl.LightOffOuterUp, map[string]interface{}{cl.Floor: i}}
			orders = append(orders, message)
		}
		if sys.OrdersDown[i] {
			sys.OrdersDown[i] = false
			message := network.Message{ "", cl.All, "", cl.LightOffOuterDown, map[string]interface{}{cl.Floor: i}}
			orders = append(orders, message)
		}
	}
	return orders
}

func (sys *System) StateChange(message *network.Message) {
	var orders []network.Message
	elevator := &sys.Elevators[message.Sender]
	switch message.Response{
	case cl.OuterOrder:
		floor := message.Content[cl.Floor]
		direction := message.Content[cl.Direction]
		if direction == 1 {
			sys.OrdersUp[floor] = true
			orders = stateChangeOrder(sys, message, cl.Up)
		} else {
			sys.OrdersDown[floor] = true
			orders = stateChangeOrder(sys, message, cl.Up)
		}
	case cl.InnerOrder:
		elevator := &sys.Elevators[message.Sender]
		floor := message.Content[cl.Floor]
		elevator.InnerOrders[floor] = true
		orders = stateChangeOrder(sys, message, cl.InnerOrder)
	case cl.Floor:
		if message.Content[cl.Floor] != -1{
			elevator.Floor = message.Content[cl.Floor]
		}
		if elevator.EngineFail {
			elevator.EngineFail = false
		}
		orders = stateChangeFloor(sys, message)
	case cl.DoorClosed:
		elevator.DoorClosed = true
		orders = stateChangeDoorClosed(sys, message)
	case cl.EngineFail:
		elevator.EngineFail = true
	}
	return orders
}

func stateChangeOrder(sys *System, message *network.Message, orderType string) []network.Message {
	var orders []network.Message
	elevator := &sys.Elevator[message.Sender]
	var elevOrders *[cl.Floors]bool
	switch {
	case cl.Up:
		elevOrders = &sys.OrdersUp
	case cl.Down:
		elevOrders = &sys.OrdersDown
	case cl.InnerOrder:
		elevOrders = &elevator.InnerOrders
	}
	if elevator.Direction == 0 {
		if elevator.Floor == message.Content[cl.Floor] {
			motor := network.Message{"", message.Sender, "", cl.Stop, ""}
			elevator.DoorClosed = false
			elevOrders[elevator.Floor] = false
			orders = append(orders, motor)
		} else if unassignedOrders(sys, message.Sender, 1) && elevator.EngineFail && elevator.DoorClosed {
			motor := network.Message{"", message.Sender, "", cl.Up, ""}
			elevator.Direction = 1
			orders = append(orders, motor)
		} else if unassignedOrders(sys, message.Sender, -1) && !elevator.EngineFail && elevator.DoorClosed {
			motor := network.Message{"", message.Sender, "", cl.Up, ""}
			elevator.Direction = -1
			orders = append(orders, motor)
		}
	}
	return orders
}

func stateChangeDoorClosed(sys *System, message *network.Message) []network.Message {
	var orders []network.Message
	elevator := &sys.Elevator[message.Sender]
	if elevator.Direction == 1 && unassignedOrders(sys, message.Sender, 1) {
		direction := network.Message{"", message.Sender, "", cl.Up, ""}
		orders = append(orders, direction)
	} else if elevator.Direction == -1 && unassignedOrders(sys, message.Sender, -1) {
		direction := network.Message{"", message.Sender, "", cl.Up, ""}
		orders = append(orders, direction)
	} else {
		elevator.Direction = 0
	}
	return orders
}

func stateChangeFloor(sys *System, message *network.Message) []network.Message {
	var orders []network.Message
	stop := false
	elevator := &sys.Elevator[message.Sender]
	if elevator.InnerOrders[elevator.Floor] {
		lightInner := network.Message{"", message.Sender, "", cl.LightOffInner, map[string]interface{}{cl.Floor: elevator.Floor}}
		elevator.InnerOrders[elevator.Floor] = false
		stop = true
		orders = append(orders, lightInner)
	}
	//This needs to get fixed
	if (sys.OrdersUp[elevator.Floor] && elevator.Direction == 1) ||
		 (unassignedOrders(sys, message.Sender, -1) && elevator.Direction == -1) {
		lightOuterUp := network.Message{"", message.Sender, "", cl.LightOffOuterUp, map[string]interface{}{cl.Floor: elevator.Floor}}
		stop = true
		sys.OrdersUp[elevator.Floor] = false
		orders = append(orders, lightOuterUp)
	}
	if (sys.OrdersDown[elevator.Floor] && elevator.Direction == -1) ||
		 (unassignedOrders(sys, message.Sender, 1) && elevator.Direction == 1) {
		lightOuterDown := network.Message{"", message.Sender, "", cl.LightOffOuterDown, map[string]interface{}{cl.Floor: elevator.Floor}}
		stop = true
		sys.OrdersDown[elevator.Floor] = false
		orders = append(orders, lightOuterDown)
	}
	if stop {
		motor := network.Message{"", message.Sender, "", cl.Stop, ""}
		elevator.DoorClosed = false
		orders = append(orders, motor)
	}
	return orders
}

func unassignedOrders(sys *System, elevIP string, dir int) bool {
	elevator := &sys.Elevators[elevIP]
	floor := elevator.Floor
	if (floor == 0 && dir == -1) || (floor == (cl.Floors - 1) && dir == 1) || dir == 0 {
		return false
	}
	if unassignedInnerOrders(elevator, floor, dir) ||
		 unassingedOuterOrder(sys, floor, elevIP, dir, cl.Up) ||
		 unassingedOuterOrder(sys, floor, elevIP, dir, cl.Down) {
		return true
	}
	return false
}

func unassingedOuterOrder(sys *System, floor int, currentIP string, dir int, orderType string) bool {
	var outerOrders [cl.Floors]bool
	switch orderType{
	case cl.Up:
		outerOrders = sys.OrdersUp
	case cl.Down:
		outerOrders = sys.OrdersDown
	}
	if dir == 1 {
		existsOrders := false
		isClosest := true
		closestFloor := floor
		for i = floor + 1; i < cl.Floors; i++ {
			existsOrders = (existsOrders || outerOrders[i])
			if existsOrders {
				closestFloor = i
			}
		}
		for elevIP, elevState := range sys.Elevators {
			if orderType == cl.Up && elevIP != currentIP && elevState.Floor > floor {
				if elevState.Floor < closestFloor &&
					 elevState.Direction == 1 {
					return false
				} else if ((elevState.Floor - closestFloor) < abs(floor - closestFloor)) &&
					 elevState.Direction == -1 {
					return false
				}
			}
			if orderType == cl.Down && elevIP != currentIP && elevState.Floor > floor {
				if elevState.Floor < closestFloor &&
					 elevState.Direction == 1 {
					return false
				} else if ((elevState.Floor - closestFloor) < abs(floor - closestFloor)) &&
					 elevState.Direction == -1 {
					return false
				}
			}
		}
	}
}

func unassignedInnerOrders(elevator *ElevatorState, floor int, dir int) bool {
	if dir == 1 {
		for i = floor + 1; i < cl.Floors; i++ {
			if elevator.InnerOrders[i] {
				return true
			}
		}
	} else {
		for i = floor - 1; i >= 0; i-- {
			if elevator.InnerOrders[i] {
				return true
			}
		}
	}
	return false
}

func abs(val int) int {
	if val < 0 {
		return -val
	} else {
		return val
	}
}
