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
	ping := time.NewTicker(100 * time.Millisecond)
	pingTimeout := time.NewTimer(300 * time.Millisecond)
	var orders []network.Message

	var sys *scheduler.System
	if fromBackup {
		sys.Recover(nwMaster.LocalIP)
		sendOrders(sys.RefreshAllElevators(), nwSlave, sys)
	} else {
		sys = scheduler.NewSystem(nwMaster.LocalIP)
	}
	for {
		select {
		case message <- recvFromSlaves:
			switch message.Response {
			case cl.InnerOrder:
				orders = sys.AddOrder(&message)
			case cl.OuterOrder:
				orders = sys.AddOrder(&message)
			case cl.Floor:
				orders = sys.StateChange(&message)
			case cl.DoorClosed:
				orders = sys.StateChange(&message)
			case cl.EngineFail:
				orders = sys.StateChange(&message)
			}
			sendOrders(orders, nwSlave, sys)
			sys.WriteToFile()
		case message <- recvFromMasters:
			switch message.Response{
			case cl.Ping:
				newElev := sys.UpdatePingTime(message.Sender)
				if newElev && sys.IsActive {
					backup := sys.CreateBackup()
					nwMaster.Send(cl.All, cl.Master, cl.Backup, backup)
				}
			case cl.Backup:
				sys.MergeSystems(&message)
				sys.SetActive(nwMaster.LocalIP)
				sendOrders(sys.RefreshAllElevators())
				sys.WriteToFile()
			}
		case <- ping:
			nwMaster.Send(cl.All, cl.Master, cl.Ping, "")
		case <- pingTimeout:
			disconnect := sys.CheckTimeout()
			if disconnect {
				if !sys.IsActive {
					sendOrders(sys.ClearOuterOrders(), nwSlave, sys)
				}
				sys.SetActive(nwMaster.LocalIP)
				sys.WriteToFile()
			}
		}
	}
}

func sendOrders(orders []network.Message, nwSlave chan<- network.Message, sys *System) {
	if !sys.IsActive {
		return
	}
	for _, order := range orders {
		nwSlave.SendMessage(order)
	}
}
