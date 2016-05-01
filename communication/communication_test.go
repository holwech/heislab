package communication

import (
	"testing"
	"time"
	"github.com/satori/go.uuid"
	"github.com/holwech/heislab/cl"
)

func TestSend(t *testing.T) {
	receiveCh, sendCh := Init(cl.SReadPort, cl.SReadPort)
	go RunPrintMsg(receiveCh)
	count := 0
	for{
		msgID := uuid.NewV4()
		msg := ResolveMsg(GetLocalIP(), GetLocalIP(), msgID.String(), "Test", map[string]interface{}{"val": count})
		sendCh <- *msg
		time.Sleep(10 * time.Second)
		count += 1
	}
}

func TestSendAndListen(t *testing.T) {
	receiveCh, sendCh := Init(cl.SReadPort, cl.SReadPort)
	go RunPrintMsg(receiveCh)
	time.Sleep(1 * time.Second)
	count := 0
	for{
		msgID := uuid.NewV4()
		msg := ResolveMsg(GetLocalIP(), GetLocalIP(), msgID.String(), "Test", map[string]interface{}{"val": count})
		sendCh <- *msg
		time.Sleep(10 * time.Second)
		count += 1
	}
}


func TestListen(t *testing.T) {
	receiveCh, _ := Init(cl.SReadPort, cl.SReadPort)
	go RunPrintMsg(receiveCh)
	time.Sleep(1 * time.Second)
	count := 0
	for{
		time.Sleep(10 * time.Second)
		count += 1
	}
}

func RunPrintMsg(receiveCh <-chan CommData) {
	for{
		PrintMessage(<- receiveCh)
	}
}

