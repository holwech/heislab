package master

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/scheduler"
	"time"
)

//REmove pls
func InitMaster() {
	go Run()
}

func Run() {
	fmt.Println("Starting master")
	nwSlave, _ := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	recvFromSlave, sendToSlave := nwSlave.Channels()
	nwMaster, _ := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromMaster, sendToMaster := nwMaster.Channels()

	sys := scheduler.NewSystem()
	slaveCommands := make(chan network.Message, 100)
	isActiveMaster := false

	pinger := time.NewTicker(100 * time.Millisecond)
	checkConnected := time.NewTicker(300 * time.Millisecond)
	connectedElevators := make(map[string]bool)
	connectedElevators[nwMaster.LocalIP] = true
	sys.AddElevator(nwMaster.LocalIP)
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
			fmt.Println(isActiveMaster)
		case <-checkConnected.C:
			masterIP := nwSlave.LocalIP
			for elevIP := range connectedElevators {
				if connectedElevators[elevIP] == false {
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
			isActiveMaster = (masterIP == nwMaster.LocalIP)
			for elevIP := range connectedElevators {
				connectedElevators[elevIP] = false
			}
			connectedElevators[nwMaster.LocalIP] = true
		case message := <-recvFromMaster:
			switch message.Response {
			case cl.Ping:
				_, alreadyConnected := connectedElevators[message.Sender]
				if isActiveMaster && !alreadyConnected {
					join := network.Message{nwSlave.LocalIP, message.Sender, network.CreateID(cl.Master), cl.System, cl.JoinMaster}
					sendToSlave <- join
					backup := sys.CreateBackup()
					backup.Receiver = message.Sender
					sendToMaster <- backup
					sys.AddElevator(message.Sender)
				}
				connectedElevators[message.Sender] = true
			case cl.Backup:
				sys.Print()
				sys = scheduler.SystemFromBackup(message)
			}
		}
	}
}
