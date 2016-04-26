package network

import (
	"errors"
	"fmt"
)

type MsgQueue struct {
	list []*Message
}

func (ol *MsgQueue) next() (Message, error) {
	var message *Message
	if len(ol.list) > 0 {
		message = ol.list[0]
		ol.remove(0)
		return *message, nil
	} else {
		return *message, errors.New("No orders left in queue")
	}
}

func (ol *MsgQueue) Resend(send chan<- Message) {
	nextOrder, stop := ol.next()
	for stop == nil {
		send <- nextOrder
	}
}

func (ol *MsgQueue) Add(order *Message) {
	ol.list = append(ol.list, order)
}

func (ol *MsgQueue) Done(id string) {
	for _, val := range ol.list {
		if val.ID == id {
			//ol.remove(i)
			fmt.Println("LOLOLOLOL", len(ol.list))
			break
		}
	}
}

func (ol *MsgQueue) remove(i int) {
	if len(ol.list) == 1 {
		ol.list[0] = nil
	} else {
		copy(ol.list[i:], ol.list[i+1:])
		ol.list[len(ol.list)-1] = nil
		ol.list = ol.list[:len(ol.list)-1]
	}
}
