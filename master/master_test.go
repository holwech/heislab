package master

import (
	"testing"
	"github.com/holwech/heislab/network"
)

func TestRun(t *testing.T) {
	slaveSend := make (chan  network.Message)
	masterSend := make( chan  network.Message)
	nw := new(network.Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)

}