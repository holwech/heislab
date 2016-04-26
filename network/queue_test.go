package network

import (
	"time"
	"testing"
	"fmt"
)

func TestAddRemove(t *testing.T) {
	queue := new(MsgQueue)
	queue.Add(&Message{"", "", "", "Num", 1})
	queue.remove(0)
}

func TestOneNext(t *testing.T) {
	queue := new(MsgQueue)
	for i := 0; i < 10; i++ {
		msg := Message{"", "", "", "Num", i}
		queue.Add(&msg)
	}
	msg, _ := queue.next()
	PrintMessage(&msg)
}

func TestAllNext(t *testing.T) {
	queue := new(MsgQueue)
	for i := 0; i < 10; i++ {
		msg := Message{"", "", "", "Num", i}
		queue.Add(&msg)
	}
	next, err := queue.next()
	for err == nil {
		fmt.Println("List length: ", len(ol.list))
		PrintMessage(&next)
		fmt.Println("List length: ", len(ol.list))
		next, err = queue.next()
	}
}

func TestResend(t *testing.T) {
	receive := make(chan Message)
	go receiverQueue(receive)
	time.Sleep(time.Second)
}

func receiverQueue(recv chan Message) {
	for {
		fmt.Println(<- recv)
	}
}
