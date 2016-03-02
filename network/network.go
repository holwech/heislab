package network

import (
	"github.com/holwech/heislab/communication"
)

type Message struct {
	Sender, Receiver, ID, Response, Content string
}
type Network struct {
	slaveReceive, slaveStatus, slaveSend, masterReceive, masterStatus, masterSend chan Message
}

func (nw *Network)Init(slaveSend chan Message, masterSend chan Message) {
	nw.slaveReceive := make(chan Message)
	nw.slaveStatus := make(chan Message)
	nw.slaveSend := slaveSend
	nw.masterReceive := make(chan Message)
	nw.masterStatus := make(chan Message)
	nw.masterSend := masterSend
}

func (nw *Network) SChannels() (<- chan Message, <- chan Message){
	return nw.slaveReceive, nw.slaveStatus
}


func (nw *Network) MChannels() (<- chan Message, <- chan Message){
	return nw.masterReceive, nw.masterStatus
}


func Run(nw *Network) {
	commSend := make(chan CommData)
	commReceive, commStatus := communication.Run()
	go sorter(nw, commSend, commReceive, commStatus)
}


func sorter(nw *Network, commSend chan<- communication.CommData, commReceive <-chan communication.CommData, commStatus <-chan communication.ConnData) {
	for{
		select{
		case message <- nw.slaveSend:
		case message <- nw.masterSend:
		}
	}
}
