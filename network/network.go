package network

import (
	"github.com/holwech/heislab/communication"
	"github.com/satori/go.uuid"
	"fmt"
)

type Message struct {
	Sender, Receiver, ID, Response string
	Content interface{}
}

type Network struct {
	slaveReceive, slaveStatus, slaveSend, masterReceive, masterStatus, masterSend chan Message
	LocalIP string
}

func PrintMessage (message *Message) {
	fmt.Printf("NETW: Sender: %s\n", message.Sender)
	fmt.Printf("NETW: Receiver: %s\n", message.Receiver)
	fmt.Printf("NETW: ID: %s\n", message.ID)
	fmt.Printf("NETW: Response: %s\n", message.Response)
	fmt.Printf("NETW: Content: %v\n", message.Content)
}

func CreateID(senderType string) (string) {
	id := uuid.NewV4()
	if senderType == "Master" {
		return  "M" + id.String()
	} else{
		return  "S" + id.String()
	}
}

func LocalIP() string {
	return communication.GetLocalIP()
}

func (nw *Network) Init(slaveSend chan Message, masterSend chan Message) {
	nw.slaveReceive = make(chan Message)
	nw.slaveStatus = make(chan Message)
	nw.slaveSend = slaveSend
	nw.masterReceive = make(chan Message)
	nw.masterStatus = make(chan Message)
	nw.masterSend = masterSend
}

func (nw *Network) SChannels() (<- chan Message, <- chan Message){
	return nw.slaveReceive, nw.slaveStatus
}


func (nw *Network) MChannels() (<- chan Message, <- chan Message){
	return nw.masterReceive, nw.masterStatus
}


func Run(nw *Network) {
	commSend := make(chan communication.CommData)
	commReceive, commStatus := communication.Run(commSend)
	nw.LocalIP = communication.GetLocalIP()
	go sorter(nw, commSend, commReceive, commStatus)
}


func sorter(nw *Network, commSend chan<- communication.CommData, commReceive <-chan communication.CommData, commStatus <-chan communication.ConnData) {
	for{
		select{
		case message := <- nw.slaveSend:
			commMsg := *communication.ResolveMsg(message.Receiver, message.ID, message.Response, message.Content)
			commSend <- commMsg
			fmt.Println("asdf")

		case message := <- nw.masterSend:
			commMsg := *communication.ResolveMsg(message.Receiver, message.ID, message.Response, message.Content)
			commSend <- commMsg
		case message := <- commReceive:
			convMsg := commToMsg(&message)
			if convMsg.Receiver == ( nw.LocalIP ) {
				nw.slaveReceive <- convMsg
			}
			if convMsg.ID[0]  != 'M'{
				nw.masterReceive <- convMsg
			}
		case status := <- commStatus:
			convStatus := connToMsg(&status)
			if convStatus.ID[0] == 'S'{
				nw.slaveStatus <- convStatus
			} else{
				nw.masterStatus <- convStatus
			}
		}
	}
}

func commToMsg(message *communication.CommData) (Message){
	newMsg := Message{
		Sender: message.SenderIP,
		Receiver: message.ReceiverIP,
		ID: message.MsgID,
		Response: message.DataType,
		Content: message.DataValue,
	}
	return newMsg
}

func connToMsg(message *communication.ConnData) (Message) {
	newMsg := Message{
		Sender: message.SenderIP,
		Receiver: "Unknown (For now anyway, i think? Maybe not)",
		ID: message.MsgID,
		Response: "Connection",
		Content: message.Status,
	}
	return newMsg
}
