package master

import (
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/orders"	
)


//Listen to inputs from slaves and send actions back
	//When do we send new orders to elevators?
	//Does activation message come from slave?
func Run(nw network.Network, sendMaster chan network.Message){
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false

	for{
	select{
	case message := <- messageChan:		
		switch message.Response{
		case cl.InnerOrder:
			sys.AddInnerOrder(message.Sender, int(message.Content))
		case cl.OuterOrder:
			sys.AddOuterOrder(int(message.Content)[0],int(message.Content)[1])
		case cl.Floor:
			sys.UpdateFloor(message.Sender,int(message.Content))
		}
	case connStatus := <- statusChan:
		switch connStatus.Response{
		case cl.Startup:
			sys.AddElevator(connStatus.Sender)
		case cl.Timeout:
			sys.RemoveElevator(connStatus.Sender)
		}
	default:
		if isActive{
			cmd, hasCommand := sys.Command()
			if hasCommand{
				sendMaster <- cmd
			}
		}
	}
	}
}