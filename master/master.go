package master

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/scheduler"
	"time"
)

func Run(backup bool) {
	fmt.Println("fmt")

	nwSlave := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	recvFromSlaves := nwSlave.Channels()
	nwMaster := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromMasters := nwMaster.Channels()

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
			case cl.EngineFail:
				sys.NotifyEngineFail(message.Sender)
			}
			sys.AssignOuterOrders()
			sys.CommandConnectedElevators(slaveCommands)
			sys.Print()
		case command := <-slaveCommands:
			if isActiveMaster {
				nwSlave.Send(command.Receiver, cl.Master, command.Response, command.Content)
			}
		case <-pingAlive.C:
			nwMaster.Send(cl.All, cl.Master, cl.Ping, "")
		case <-checkConnected.C:
			for elevatorIP := range connectedElevators {
				if connectedElevators[elevatorIP] == false {
					if isActiveMaster {
						sys.NotifyDisconnectionActive(elevatorIP)
					} else {
						sys.NotifyDisconnectionInactive(elevatorIP)
					}
					fmt.Println(elevatorIP, " disconnected")
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
					merge := sys.ToMessage()
					for elevatorIP := range connectedElevators {
						merge.Receiver = elevatorIP
						nwMaster.SendMessage(merge)
					}
				}

			case cl.Backup:
				receivedSys := scheduler.SystemFromMessage(message)
				sys = scheduler.MergeSystems(sys, receivedSys)
			}
		}
	}
}
