package network

import (
	"fmt"

	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/communication"
	"github.com/satori/go.uuid"
)

const info = true

type Message struct {
	Sender, Receiver, ID, Response string
	Content                        interface{}
}

type Network struct {
	slaveReceive, slaveSend, masterReceive, masterSend chan Message
	LocalIP                                            string
}

func printInfo(comment string, message *Message) {
	if info && message.Response != cl.Connection {
		fmt.Println("NETW: " + comment)
		PrintMessage(message)
	}
}

func PrintMessage(message *Message) {
	fmt.Printf("NETW: Sender: %s\n", message.Sender)
	fmt.Printf("NETW: Receiver: %s\n", message.Receiver)
	fmt.Printf("NETW: ID: %s\n", message.ID)
	fmt.Printf("NETW: Response: %s\n", message.Response)
	fmt.Printf("NETW: Content: %v\n", message.Content)
}

func CreateID(senderType string) string {
	id := uuid.NewV4()
	if senderType == cl.Master {
		return "M" + id.String()
	} else if senderType == cl.Slave {
		return "S" + id.String()
	} else {
		return "=== ERROR: Wrong input in network.CreateID"
	}
}

func LocalIP() string {
	return communication.GetLocalIP()
}

func (nw *Network) Init(slaveSend chan Message, masterSend chan Message) {
	nw.slaveReceive = make(chan Message)
	nw.slaveSend = slaveSend
	nw.masterReceive = make(chan Message)
	nw.masterSend = masterSend
}

func (nw *Network) SChannels() <-chan Message {
	return nw.slaveReceive
}

func (nw *Network) MChannels() <-chan Message {
	return nw.masterReceive
}

func Run(nw *Network) {
	commSend := make(chan communication.CommData)
	commReceive := communication.Run(commSend)
	nw.LocalIP = communication.GetLocalIP()
	go sorter(nw, commSend, commReceive)
}

func sorter(nw *Network, commSend chan<- communication.CommData, commReceive <-chan communication.CommData) {
	for {
		select {
		case message := <-nw.slaveSend:
			commMsg := *communication.ResolveMsg(nw.LocalIP, message.Receiver, message.ID, message.Response, message.Content)
			commSend <- commMsg
		case message := <-nw.masterSend:
			commMsg := *communication.ResolveMsg(nw.LocalIP, message.Receiver, message.ID, message.Response, message.Content)
			commSend <- commMsg
		case message := <-commReceive:
			convMsg := commToMsg(&message)
			if ((convMsg.Response != cl.Connection) && (convMsg.ID[0] == 'M')) ||
				((convMsg.Response == cl.Connection) && (convMsg.ID[0] == 'S')) {
				nw.slaveReceive <- convMsg
				printInfo("Slave received message", &convMsg)
			}
			if ((convMsg.ID[0] == 'S') && (convMsg.Response != cl.Connection)) ||
				((convMsg.ID[0] == 'M') && (convMsg.Response == cl.Connection)) {
				nw.masterReceive <- convMsg
				printInfo("Master received message", &convMsg)
			}
		}
	}
}

func commToMsg(message *communication.CommData) Message {
	newMsg := Message{
		Sender:   message.SenderIP,
		Receiver: message.ReceiverIP,
		ID:       message.MsgID,
		Response: message.Response,
		Content:  message.Content,
	}
	return newMsg
}
