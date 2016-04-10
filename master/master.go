package master

import (
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/orders"	
	"fmt"
)

func Init(nw *network.Network, sendMaster chan network.Message) {
	go Run(nw,sendMaster)
}
//Listen to inputs from slaves and send commands back
	//Will the behaviour and order list be the same on all masters running?
func Run(nw *network.Network, sendMaster chan network.Message){
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false
	for{
		select{
		case message := <- messageChan:
			systemChange := false
			switch message.Response{
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})	
				floor := int(content["Floor"].(float64))
				sys.AssignOrder(message.Sender,floor)
				systemChange = true			
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})	
				floor := int(content["Floor"].(float64))
				direction := int(content["Direction"].(float64))
				sys.AddOuterOrder(floor,direction)
				systemChange = true
			case cl.Floor:		
				floor := int(message.Content.(float64))
				cmd, hasCommand := sys.FloorAction(message.Sender,floor)
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive{
					sendMaster <- cmd
				}
				systemChange = true
			case cl.DoorClosed:
				sys.DoorClosedEvent(message.Sender)
				systemChange = true
			case cl.Startup:
				ping := network.Message{network.LocalIP(),message.Sender,network.CreateID(cl.Master),cl.JoinMaster,""}
				if isActive{
					sendMaster <- ping
				}
				sys.AddElevator(message.Sender)
			case cl.SetMaster:
				isActive = true
			}
			if systemChange{
				sys.AssignOrders()
				cmd, hasCommand := sys.CommandElevators()
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive{
					sendMaster <- cmd
				}
			}
		case connStatus := <- statusChan:
			switch connStatus.Response{
			case cl.Timeout:
				sys.RemoveElevator(connStatus.Sender)
			}
		}
	}
}