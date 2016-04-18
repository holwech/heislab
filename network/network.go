package network

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/communication"
	"github.com/satori/go.uuid"
)

const info = true
const printAll = false
const conn = false

type Message struct {
	Sender, Receiver, ID, Response string
	Content                        interface{}
}

type Network struct {
	slaveReceive, slaveSend, masterReceive, masterSend chan Message
	LocalIP                                            string
}

func printInfo(comment string, message *Message) {
	if (info && message.Response != cl.Connection) || conn {
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

func printError(errMsg string, err error) {
	fmt.Printf(errMsg + "\n")
	fmt.Printf(err.Error() + "\n")
	fmt.Println()
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

func InitNetwork() *Network {
	nw := new(Network)
	nw.Init()
	Run(nw)
	return nw
}

func (nw *Network) Init() {
	nw.slaveReceive = make(chan Message)
	nw.masterReceive = make(chan Message)
	nw.slaveSend = make(chan Message)
	nw.masterSend = make(chan Message)
}

func (nw *Network) SChannels() (<-chan Message, chan<- Message) {
	return nw.slaveReceive, nw.slaveSend
}

func (nw *Network) MChannels() (<-chan Message, chan<- Message) {
	return nw.masterReceive, nw.masterSend
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
			assertMsg(&convMsg)
			if printAll {
				PrintMessage(&convMsg)
			}
			if convMsg.Response != cl.Connection && convMsg.ID[0] == 'M' &&
				(convMsg.Receiver == nw.LocalIP || convMsg.Receiver == cl.All) ||
				(convMsg.Response == cl.Connection && convMsg.ID[0] == 'S') {
				nw.slaveReceive <- convMsg
				printInfo("Slave received message", &convMsg)
			}
			if ((convMsg.ID[0] == 'S') && (convMsg.Response != cl.Connection)) ||
				((convMsg.ID[0] == 'M') && (convMsg.Response == cl.Connection) &&
					(convMsg.Receiver == nw.LocalIP)) {
				nw.masterReceive <- convMsg
				printInfo("Master received message", &convMsg)
			}
		}
	}
}

func assertMsg(message *Message) {
	switch message.Content.(type) {
	case float64:
		message.Content = int(message.Content.(float64))
	case map[string]interface{}:
		tempMap := message.Content.(map[string]interface{})
		for key, value := range tempMap {
			switch value.(type) {
			case float64:
				tempMap[key] = int(value.(float64))
			}
		}
		message.Content = tempMap
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
