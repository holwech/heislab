package scheduler

import (
	"testing"
	"time"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

func TestBackupOverNetwork(t *testing.T){
	nw_M := network.InitNetwork(cl.MReadPort, cl.MWritePort, cl.Master)
	nw_S := network.InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	time.Sleep(100*time.Millisecond)
	receive_M, send_M := nw_M.Channels()
	receive_S, send_S := nw_S.Channels()

	sys := NewSystem()
	sys.AddElevator(nw_S.LocalIP)
	sys.NotifyInnerOrder(nw_S.LocalIP,3)
	sys.NotifyOuterOrder(2,1)
	sys.Print()
	b := sys.CreateBackup()
	send_M <- b
	send_S <-<-receive_S
	sys2 := SystemFromBackup(<- receive_M)
	sys2.Print()
}