package master

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"github.com/holwech/heislab/scheduler"
	"time"
)


func Run(fromBackup bool) {
	nwSlave := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	nwMaster := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromSlaves := nwSlave.Channels()
	recvFromMasters := nwMaster.Channels()
	slaveCommands := make(chan network.Message, 100)

}
func Run(fromBackup bool) {
	nwSlave := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	nwMaster := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromSlaves := nwSlave.Channels()
	recvFromMasters := nwMaster.Channels()
	slaveCommands := make(chan network.Message, 100)
func Run(fromBackup bool) {
	nwSlave := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	nwMaster := network.InitNetwork(cl.MtoMPort, cl.MtoMPort, cl.Master)
	recvFromSlaves := nwSlave.Channels()
	recvFromMasters := nwMaster.Channels()
	slaveCommands := make(chan network.Message, 100)

	time.Sleep(50 * time.Millisecond)
	sys := scheduler.NewSystem()
	if fromBackup {
		sys = scheduler.ReadFromFile()
		elevator := sys.Elevators[nwMaster.LocalIP]
		elevator.CurrentBehaviour = scheduler.Idle
		if elevator.HasMoreOrders() {
			elevator.AwaitingCommand = true
		}
		sys.Elevators[nwMaster.LocalIP] = elevator
	} else {
		sys.AddElevator(nwMaster.LocalIP)
	}
	sys.Print()

	connectedElevators := make(map[string]bool)
	connectedElevators[nwMaster.LocalIP] = true
	pingAlive := time.NewTicker(75 * time.Millisecond)
	checkConnected := time.NewTicker(300 * time.Millisecond)
	isActiveMaster := true

	for {
		select {
		case message := <-recvFromSlaves:
			switch message.Response {
			case cl.InnerOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				command := sys.NotifyInnerOrder(message.Sender, floor)
				slaveCommands <- <-command
			case cl.OuterOrder:
				content := message.Content.(map[string]interface{})
				floor := content["Floor"].(int)
				direction := content["Direction"].(int)
				command := sys.NotifyOuterOrder(floor, direction)
				slaveCommands <- <-command
			case cl.Floor:
				floor := message.Content.(int)
				sys.NotifyFloor(message.Sender, floor)
			case cl.DoorClosed:
				sys.NotifyDoorClosed(message.Sender)
			case cl.EngineFail:
				sys.NotifyEngineFail(message.Sender)
			}
			sys.AssignOuterOrders()
			commands := sys.CommandConnectedElevators()
			for command := range commands {
				slaveCommands <- command
			}
			sys.Print()
			go sys.WriteToFile()
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
					commands := sys.CommandConnectedElevators()
					for command := range commands {
						slaveCommands <- command
					}
					go sys.WriteToFile()
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
