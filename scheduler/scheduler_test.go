package scheduler

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"testing"
	"time"
)

func TestBackupOverNetwork(t *testing.T) {
	nw_M := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	nw_S := network.InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	time.Sleep(100 * time.Millisecond)
	receive_M, send_M := nw_M.Channels()
	receive_S, send_S := nw_S.Channels()

	commandsToSlave := make(chan network.Message, 100)

	sys := NewSystem()
	sys.AddElevator(nw_S.LocalIP)
	sys.NotifyInnerOrder(nw_S.LocalIP, 3, commandsToSlave)
	sys.NotifyOuterOrder(2, 1, commandsToSlave)
	sys.Print()
	b := sys.CreateBackup()
	b.Receiver = cl.All
	send_M <- b
	send_S <- <-receive_S

	sys2 := SystemFromBackup(<-receive_M)
	sys2.Print()
}
