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
	nwSlave := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	recvFromSlave, sendToSlave := nwSlave.Channels()
	nwMaster := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
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
		case <-checkConnected.C:
			for elevIP := range connectedElevators {
				if connectedElevators[elevIP] == false {
					if isActiveMaster {
						sys.NotifyDisconnectionActive(elevIP)
					} else {
						sys.NotifyDisconnectionInactive(elevIP)
					}
					delete(connectedElevators, elevIP)
				}
			}
			masterIP := nwSlave.LocalIP
			for elevIP := range connectedElevators {
				if elevIP < masterIP {
					masterIP = elevIP
				}
				connectedElevators[elevIP] = false
			}
			isActiveMaster = (masterIP == nwMaster.LocalIP)
			connectedElevators[nwMaster.LocalIP] = true
		case message := <-recvFromMaster:
			switch message.Response {
			case cl.Ping:
				_, alreadyConnected := connectedElevators[message.Sender]
				if isActiveMaster && !alreadyConnected {
					backup := sys.CreateBackup()
					backup.Receiver = message.Sender
					sendToMaster <- backup
				}
				connectedElevators[message.Sender] = true
			case cl.Backup:
				sys.Print()
				sys2 := scheduler.SystemFromBackup(message)
				sys = scheduler.MergeSystems(sys, sys2)
			}
		}
	}
}
