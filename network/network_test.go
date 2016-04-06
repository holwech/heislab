package communication

import (
	"testing"
	"time"
	"github.com/satori/go.uuid"
)

func TestSend(t *testing.T) {
	slaveSend := make(chan network.Message)
	slaveMaster := make(chan network.Message)
	nw := new(network.Network)
	nw.Init(slaveSend, masterSend)
	network.Run(nw)
}

