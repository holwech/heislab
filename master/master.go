package master

import (
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/orders"	
)


//Listen to inputs from slaves and send actions back
	//When do we send new orders to elevators?
	//Does activation message come from slave?
	//Will the behaviour and order list be the same on all masters running?
func Run(nw network.Network, sendMaster chan network.Message){
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false

	for{
	select{
	case message := <- messageChan:		
		switch message.Response{
		case cl.InnerOrder:
			sys.AddInnerOrder(message.Sender,message.Content.(int))
		case cl.OuterOrder:
			sys.AddOuterOrder(int(message.Content.(string)[0]),int(message.Content.(string)[1]))
		case cl.Floor:
			sys.UpdateFloor(message.Sender,message.Content.(int))
		case cl.Startup:
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
	default:
			cmd, hasCommand := sys.Command()
			cmd.Sender = network.LocalIP()
			cmd.ID = network.CreateID(cl.Master)
			if hasCommand && isActive{
				sendMaster <- cmd
			}
		}
	}
}