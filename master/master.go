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
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false
	for {
		select {
		case message := <-messageChan:
			systemChange := false
			switch message.Response {
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})
				floor := int(content["Floor"].(float64))
				systemChange = systemChange || sys.AddInnerOrder(message.Sender, floor)
				lightCmd := network.Message{network.LocalIP(), message.Sender, network.CreateID(cl.Master), cl.LightOnInner, floor}
				sendMaster <- lightCmd
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})
				floor := int(content["Floor"].(float64))
				direction := int(content["Direction"].(float64))
				systemChange = systemChange || sys.AddOuterOrder(floor, direction)
				if direction == 1 {
					lightCmd := network.Message{network.LocalIP(), cl.All, network.CreateID(cl.Master), cl.LightOnOuterUp, floor}
					sendMaster <- lightCmd
				} else {
					lightCmd := network.Message{network.LocalIP(), cl.All, network.CreateID(cl.Master), cl.LightOnOuterDown, floor}
					sendMaster <- lightCmd
				}
			case cl.Floor:
				floor := int(message.Content.(float64))
				cmd, hasCommand := sys.FloorAction(message.Sender, floor)
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive {
					sendMaster <- cmd
				}
				systemChange = systemChange || hasCommand
			case cl.DoorClosed:
				sys.DoorClosedEvent(message.Sender)
				systemChange = true
			case cl.Startup:
				ping := network.Message{network.LocalIP(), message.Sender, network.CreateID(cl.Master), cl.JoinMaster, ""}
				if isActive {
					sendMaster <- ping
				}
				sys.AddElevator(message.Sender)
			case cl.SetMaster:
				isActive = true
			}
			if systemChange {
				sys.AssignOrders()
				cmd, hasCommand := sys.GetNextCommand()
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive {
					sendMaster <- cmd
				}
			}
			fmt.Println(sys)

		case connStatus := <-statusChan:
			switch connStatus.Response {
			case cl.Timeout:
				sys.RemoveElevator(connStatus.Sender)
			}
		}
	}
}
