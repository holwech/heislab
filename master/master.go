package master

import (
	"fmt"

	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/orders"
)

func InitMaster(nw *network.Network) {
	go Run(nw)
}

//Listen to inputs from slaves and send commands back
//Will the behaviour and order list be the same on all masters running?
func Run(nw *network.Network) {
	inputChan, sendMaster := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false
	for {
		select {
		case message := <-inputChan:
			switch message.Response {
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				sys.AddInnerOrder(message.Sender, floor)
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				direction := content["Direction"].(int)
				sys.AddOuterOrder(floor, direction)

			case cl.Floor:
				floor := message.Content.(int)
				sys.NotifyFloor(message.Sender, floor)

			case cl.DoorClosed:
				sys.NotifyDoorClosed(message.Sender)

			case cl.Timeout:
				//Future work - check connected elevators
				sys.RemoveElevator(message.Sender)
			case cl.Startup:
				if isActive {
					ping := network.Message{nw.LocalIP, message.Sender, network.CreateID(cl.Master), cl.JoinMaster, ""}
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
				message.Sender = nw.LocalIP
				message.ID = network.CreateID(cl.Master)
				sendMaster <- message
			}
		}
	}
}
