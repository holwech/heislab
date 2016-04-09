package master

import (
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/orders"	
	"time"
)

func Init(nw *network.Network, sendMaster chan network.Message) {
	go Run(nw,sendMaster)
}
//Listen to inputs from slaves and send actions back
	//When do we send new orders to elevators?
	//Does activation message come from slave?
	//Will the behaviour and order list be the same on all masters running?
func Run(nw *network.Network, sendMaster chan network.Message){
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := true
	ticker := time.NewTicker(50 * time.Millisecond)
	sys.AddElevator("129.241.187.146")
	for{
	select{
	case message := <- messageChan:	
		switch message.Response{
		case cl.InnerOrder:
			content := message.Content.(map[string]interface{})	
			floor := int(content["Floor"].(float64))
			sys.AssignOrder(message.Sender,floor)			
		case cl.OuterOrder:
			content := message.Content.(map[string]interface{})	
			floor := int(content["Floor"].(float64))
			direction := int(content["Direction"].(float64))
			sys.AddOuterOrder(floor,direction)
		case cl.Floor:
			floor := int(message.Content.(float64))
			sys.UpdateFloor(message.Sender,floor)
		case cl.Startup:
			ping := network.Message{network.LocalIP(),message.Sender,network.CreateID(cl.Master),cl.JoinMaster,""}
			sendMaster <- ping
			sys.AddElevator(message.Sender)
		case cl.Ping:
			ping := network.Message{network.LocalIP(),message.Sender,network.CreateID(cl.Master),cl.Ping,""}
			sendMaster <- ping
		case cl.SetMaster:
			isActive = true
		}
	case connStatus := <- statusChan:
		switch connStatus.Response{
		case cl.Timeout:
			sys.RemoveElevator(connStatus.Sender)
	}
	case <-ticker.C:
			cmd, hasCommand := sys.Command()
			cmd.Sender = network.LocalIP()
			cmd.ID = network.CreateID(cl.Master)
			if hasCommand && isActive{
				sendMaster <- cmd
			}
		}
	}
}