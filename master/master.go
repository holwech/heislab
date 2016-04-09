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
//Listen to inputs from slaves and send actions back
	//When do we send new orders to elevators?
	//Does activation message come from slave?
	//Will the behaviour and order list be the same on all masters running?
func Run(nw *network.Network, sendMaster chan network.Message){
	messageChan, statusChan := nw.MChannels()
	sys := orders.NewSystem()
	isActive := false
	for{
		select{
		case message := <- messageChan:
			event := false
			switch message.Response{
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})	
				floor := int(content["Floor"].(float64))
				sys.AssignOrder(message.Sender,floor)
				event = true			
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})	
				floor := int(content["Floor"].(float64))
				direction := int(content["Direction"].(float64))
				sys.AddOuterOrder(floor,direction)
				event = true
			case cl.Floor:		
				floor := int(message.Content.(float64))
				cmd, hasCommand := sys.FloorAction(message.Sender,floor)
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive{
					sendMaster <- cmd
				}
				event = true
			case cl.Startup:
				ping := network.Message{network.LocalIP(),message.Sender,network.CreateID(cl.Master),cl.JoinMaster,""}
				if isActive{
					sendMaster <- ping
				}
				sys.AddElevator(message.Sender)
			case cl.DoorClosed:
				sys.DoorClosedEvent(message.Sender)
				event = true
			case cl.SetMaster:
				isActive = true
			}
			if event{
			cmd, hasCommand := sys.Command()
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
		/*case <-ticker.C:
				cmd, hasCommand := sys.Command()
				cmd.Sender = network.LocalIP()
				cmd.ID = network.CreateID(cl.Master)
				if hasCommand && isActive{
					fmt.Println("MASTE34R")
					sendMaster <- cmd
				}
		*/
		}
	}
			fmt.Println(sys)

}