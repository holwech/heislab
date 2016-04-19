package master

import (
	"testing"
	"fmt"
	"time"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
)

func TestRun(t *testing.T) {
	nw := network.InitNetwork()
	InitMaster(nw)
	network.Run(nw)
	sR, sS := nw.SChannels()
	sS <- network.Message{network.LocalIP(),network.LocalIP(),network.CreateID(cl.Slave) ,cl.Startup, time.Now()}
	msg := <- sR
	fmt.Println(msg)
}