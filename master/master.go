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

	isActiveMaster := true
	time.Sleep(50 * time.Millisecond)
	sys := scheduler.NewSystem()
	if fromBackup {
		backup := scheduler.ReadFromFile()
		var elevator scheduler.ElevatorState
		if len(backup.Elevators) == 1 {
			sys = backup
			elevator = backup.Elevators[nwMaster.LocalIP]
		} else {
			elevator = backup.Elevators[nwMaster.LocalIP]
			sys.UnhandledOrdersUp = backup.UnhandledOrdersUp
			sys.UnhandledOrdersDown = backup.UnhandledOrdersDown
		}
		elevator.CurrentBehaviour = scheduler.Idle
		elevator.Direction = 1
		elevator.Floor = 0
		if elevator.HasMoreOrders() {
			elevator.AwaitingCommand = true
		}
		sys.Elevators[nwMaster.LocalIP] = elevator
		lights := sys.SendLightCommands()
		sys.Print()
		for _, command := range lights {
			nwSlave.SendMessage(command)
		}
		commands := sys.CommandConnectedElevators()
		for command := range commands {
			slaveCommands <- command
		}
		fmt.Println("Loading backup sucessfull!")
	} else {
		sys.AddElevator(nwMaster.LocalIP)
	}
	sys.Print()

	connectedElevators := make(map[string]bool)
	connectedElevators[nwMaster.LocalIP] = true
	pingAlive := time.NewTicker(75 * time.Millisecond)
	checkConnected := time.NewTicker(300 * time.Millisecond)
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
					sys.Print()
					masterIP := nwSlave.LocalIP
					for elevatorIP := range connectedElevators {
						if elevatorIP < masterIP {
							masterIP = elevatorIP
						}
					}
					isActiveMaster = (masterIP == nwMaster.LocalIP)
					lights := sys.SendLightCommands()
					for _, command := range lights {
						slaveCommands <- command
					}
					go sys.WriteToFile()
				} else {
					connectedElevators[elevatorIP] = false
				}
			}
			connectedElevators[nwMaster.LocalIP] = true
		case message := <-recvFromMasters:
			switch message.Response {
			case cl.Ping:
				_, alreadyConnected := connectedElevators[message.Sender]
				connectedElevators[message.Sender] = true
				if !alreadyConnected {
					merge := sys.ToMessage()
					for elevatorIP := range connectedElevators {
						if elevatorIP != nwMaster.LocalIP {
							merge.Receiver = elevatorIP
							nwMaster.SendMessage(merge)
						}
					}
					fmt.Println(message.Sender, " connected")
					//Select master as connected elevator with lowest IP
					masterIP := nwSlave.LocalIP
					for elevatorIP := range connectedElevators {
						if elevatorIP < masterIP {
							masterIP = elevatorIP
						}
					}
					isActiveMaster = (masterIP == nwMaster.LocalIP)
					fmt.Println(masterIP, " is active master")
				}
			case cl.Backup:
				receivedSys := scheduler.SystemFromMessage(message)
				scheduler.MergeSystems(sys, receivedSys)
				commands := sys.SendLightCommands()
				for _, command := range commands {
					slaveCommands <- command
				}
				sys.Print()
			}
		}
	}
}
