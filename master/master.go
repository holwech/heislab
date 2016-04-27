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
	fmt.Println("fmt")

	nwSlave, _ := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	recvFromSlaves, sendToSlaves := nwSlave.Channels()
	nwMaster, _ := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromMasters, sendToMasters := nwMaster.Channels()

	sys := scheduler.NewSystem()
	slaveCommands := make(chan network.Message, 100)
	isActiveMaster := true

	pingAlive := time.NewTicker(75 * time.Millisecond)
	checkConnected := time.NewTicker(300 * time.Millisecond)
	connectedElevators := make(map[string]bool)
	connectedElevators[nwMaster.LocalIP] = true
	sys.AddElevator(nwMaster.LocalIP)

	for {
		select {
		case message := <-recvFromSlaves:
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
				fmt.Println("doorclose")
			case cl.System:
				switch message.Content {
				case cl.EngineFail:
					sys.NotifyEngineFail(message.Sender)
				}
			}
			sys.AssignOuterOrders()
			sys.CommandConnectedElevators(slaveCommands)
			sys.Print()
		case command := <-slaveCommands:
			if isActiveMaster {
				command.Sender = nwSlave.LocalIP
				command.ID = network.CreateID(cl.Master)
				sendToSlaves <- command
			}
		case <-pingAlive.C:
			ping := network.Message{nwMaster.LocalIP, cl.All, network.CreateID(cl.Master), cl.Ping, ""}
			sendToMasters <- ping
		case <-checkConnected.C:
			for elevatorIP := range connectedElevators {
				if connectedElevators[elevatorIP] == false {
					if isActiveMaster {
						sys.NotifyDisconnectionActive(elevatorIP)
					} else {
						sys.NotifyDisconnectionInactive(elevatorIP)
					}
					delete(connectedElevators, elevatorIP)
				}
			}
			//Select master as connected elevator with lowest IP
			masterIP := nwSlave.LocalIP
			for elevatorIP := range connectedElevators {
				if elevatorIP < masterIP {
					masterIP = elevatorIP
				}
				connectedElevators[elevatorIP] = false
			}
			isActiveMaster = (masterIP == nwMaster.LocalIP)
			connectedElevators[nwMaster.LocalIP] = true
		case message := <-recvFromMasters:
			switch message.Response {
			case cl.Ping:
				_, alreadyConnected := connectedElevators[message.Sender]
				connectedElevators[message.Sender] = true
				if !alreadyConnected {
					fmt.Println(connectedElevators)
					merge := sys.ToMessage()
					for elevatorIP := range connectedElevators {
						if elevatorIP != nwMaster.LocalIP {
							merge.Receiver = elevatorIP
							sendToMasters <- merge
						}
					}
				}

			case cl.Backup:
				receivedSys := scheduler.SystemFromMessage(message)
				sys = scheduler.MergeSystems(sys, receivedSys)
			}
		}
	}
}
