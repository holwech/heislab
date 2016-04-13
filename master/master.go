package master

import (
	"fmt"

	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/orders"
)

func Init(nw *network.Network, sendMaster chan network.Message) {
	go Run(nw, sendMaster)
}

//Listen to inputs from slaves and send commands back
//Will the behaviour and order list be the same on all masters running?
func Run(nw *network.Network, sendMaster chan network.Message) {
	inputChan, _ := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false
	for {
		select {
		case message := <-inputChan:
			switch message.Response {
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})
				floor := int(content["Floor"].(float64))
				sys.AddInnerOrder(message.Sender, floor)
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})
				floor := int(content["Floor"].(float64))
				direction := int(content["Direction"].(float64))
				sys.AddOuterOrder(floor, direction)

			case cl.Floor:
				floor := int(message.Content.(float64))
				sys.NotifyFloor(message.Sender, floor)

			case cl.DoorClosed:
				sys.NotifyDoorClosed(message.Sender)

			case cl.Timeout:
				//Future work - check connected elevators
				sys.RemoveElevator(message.Sender)
			case cl.Startup:
				if isActive {
					ping := network.Message{network.LocalIP(), message.Sender, network.CreateID(cl.Master), cl.JoinMaster, ""}
					sendMaster <- ping
				}
				sys.AddElevator(message.Sender)
			case cl.SetMaster:
				isActive = true
			}
			sys.AssignOrders()
			sys.CheckNewCommand()
			fmt.Println(sys)
		case message := <-sys.Commands:
			if isActive {
				message.Sender = network.LocalIP()
				message.ID = network.CreateID(cl.Master)
				sendMaster <- message
			}
		}
	}
}
