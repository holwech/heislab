package master

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/scheduler"
)

//REmove pls
func InitMaster() {
	go Run()
}

func Run() {
	nwSlave, _ := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	recvFromSlave, sendToSlave := nwSlave.Channels()
	nwMaster, _ := network.InitNetwork(cl.MtoMReadPort, cl.MtoMWritePort, cl.Master)
	recvFromMaster, sendToMaster := nwMaster.Channels()

	sys := scheduler.NewSystem()
	slaveCommands := make(chan network.Message, 100)
	isActiveMaster := false

	pinger := time.NewTicker(100 * time.Millisecond)
	checkConnected := time.NewTicker(300 * time.Millisecond)
	connectedElevators = make(map[string]bool)
	for {
		select {
		case message := <-recvFromSlave:
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
			case cl.System:
				switch message.Content {
				case cl.EngineFail:
					sys.NotifyEngineFail(message.Sender)
				case cl.EngineOK:
					sys.NotifyEngineOk(message.Sender)
				}
			}
			sys.AssignOuterOrders()
			sys.CommandConnectedElevators(slaveCommands)
			sys.Print()
		case command := <-slaveCommands:
			if isActiveMaster {
				command.Sender = nwSlave.LocalIP
				command.ID = network.CreateID(cl.Master)
				sendToSlave <- command
			}
		case <-pinger.C:
			ping := network.Message{nwMaster.LocalIP, cl.All, network.CreateID(cl.Master), cl.Ping, ""}
			sendToMaster <- ping
		case <-checkConnected.C:
			masterIP := nw.LocalIP
			for elevIP := range connectedElevators {
				if checkConnected[elevIP] == false {
					if isActiveMaster {
						sys.NotifyDisconnectionActive(elevIP)
					} else {
						sys.NotifyDisconnectionInactive(elevIP)
					}
				} else {
					//Select master as the connected elevator with lowest IP
					if elevIP < masterIP {
						masterIP = elevIP
					}
				}
			}
			isActiveMaster = (masterIP == nw.LocalIP)

			for elevIP := range connectedElevators {
				checkConnected[elevIP] = false
			}
		case message := <-recvFromMaster:
			switch message.Response {
			case cl.Ping:
				_, elevatorAdded := connectedElevators[message.Sender]
				if isActiveMaster && !elevatorAdded {
					join := network.Message{nwSlave.LocalIP, message.Sender, network.CreateID(cl.Master), cl.System, cl.JoinMaster}
					sendToSlave <- join
					backup := sys.CreateBackup()
					backup.Receiver = message.Sender
					sendToMaster <- backup
					sys.AddElevator(message.Sender)
				}
				connectedElevators[message.Sender] = true
			case cl.Backup:
				sys = scheduler.SystemFromBackup(message)
			}
		}
	}
}
