package scheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"io/ioutil"
)

//TODO: Make this actually work right
func (sys *System) CalculateOrders() []network.Message {
	var orders map[string]int
	
}

func calculateOrdersUp() 


func (sys *System) AddOrder(message *network.Message) {
	switch message.Response {
	case cl.OuterOrder:
		floor := message.Content[cl.Floor]
		direction := message.Content[cl.Direction]
		if direction == 1 {
			sys.OrdersUp[floor] = true
		} else {
			sys.OrdersDown[floor] = true
		}
	case cl.InnerOrder:
		elevator := &sys.Elevators[message.Sender]
		floor := message.Content[cl.Floor]
		elevator.InnerOrders[floor] = true
	}
}

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
			message := network.Message{ "", cl.All, "", cl.LightOffOuterOff, map[string]interface{}{cl.Floor: i}}
			orders = append(orders, message)
		}
	}
	return orders
}

func (sys *System) StateChange(message *network.Message) {
	elevator := &sys.Elevators[message.Sender]
	switch message.Response{
	case cl.Floor:
		if message.Content[cl.Floor] != -1{
			elevator.Floor = message.Content[cl.Floor]
		}
		if elevator.EngineFail {
			elevator.EngineFail = false
		}
	case cl.DoorClosed:
		elevator.DoorClosed = true
	case cl.EngineFail:
		elevator.EngineFail = true
	}
}


