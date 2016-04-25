package master

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/scheduler"
)

func InitMaster() {
	go Run()
}

//Listen to inputs from slaves and send commands back
//Will the behaviour and order list be the same on all masters running?
func Run() {
	nw, ol := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	receive, send := nw.Channels()
	sys := scheduler.NewSystem()
	slaveCommands := make(chan network.Message, 100)
	isActiveMaster := false

	for {
		select {
		case message := <-receive:
			switch message.Response {
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				sys.NotifyInnerOrder(message.Sender, floor, slaveCommands)

			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				direction := content["Direction"].(int)
				sys.NotifyOuterOrder(floor, direction, slaveCommands)

			case cl.Floor:
				floor := message.Content.(int)
				sys.NotifyFloor(message.Sender, floor, slaveCommands)

			case cl.DoorClosed:
				sys.NotifyDoorClosed(message.Sender)

			case cl.Connection:
				sys.RemoveElevator(message.Sender)
			case cl.Backup:
				sys = scheduler.SystemFromBackup(message)
			case cl.System:
				switch message.Content {
				case cl.Startup:
					if isActiveMaster {
						ping := network.Message{nw.LocalIP, message.Sender, network.CreateID(cl.Master), cl.System, cl.JoinMaster}
						send <- ping
					}
					sys.AddElevator(message.Sender)
					backup := sys.CreateBackup()
					backup.Receiver = message.Sender
					send <- backup
				case cl.SetMaster:
					isActiveMaster = true
				case cl.EngineFail:
					sys.NotifyEngineFail(message.Sender)
				case cl.EngineOK:
					sys.NotifyEngineOk(message.Sender)
				}
			case cl.Connection:
				ol.Done(message.ID)
			}
			sys.AssignOuterOrders()
			sys.CommandConnectedElevators(slaveCommands)
			sys.Print()
		case command := <-slaveCommands:
			if isActiveMaster {
				command.Sender = nw.LocalIP
				command.ID = network.CreateID(cl.Master)
				send <- command
			}
		}

	}
}
