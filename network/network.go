package network

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/communication"
	"github.com/satori/go.uuid"
)

const info = true
const conn = false

type Message struct {
	Sender, Receiver, ID, Response string
	Content                        interface{}
}

type Network struct {
	Receive, Send                            chan Message
	SenderType, LocalIP, ReadPort, WritePort string
}

func (nw *Network) Init(readPort string, writePort string, senderType string) {
	nw.Receive = make(chan Message)
	nw.Send = make(chan Message)
	nw.LocalIP = communication.GetLocalIP()
	nw.ReadPort = readPort
	nw.WritePort = writePort
	nw.SenderType = senderType
}

func (nw *Network) Channels() (<-chan Message, chan<- Message) {
	return nw.Receive, nw.Send
}

func InitNetwork(readPort string, writePort string, senderType string) *Network {
	nw := new(Network)
	nw.Init(readPort, writePort, senderType)
	commReceive, commSend := communication.Init(nw.ReadPort, nw.WritePort)
	go receiver(nw, commReceive)
	go sender(nw, commSend)
	return nw
}

func sender(nw *Network, commSend chan<- communication.CommData) {
	for {
		message := <-nw.Send
		commMsg := *communication.ResolveMsg(nw.LocalIP, message.Receiver, message.ID, message.Response, message.Content)
		commSend <- commMsg
	}
}

func receiver(nw *Network, commReceive <-chan communication.CommData) {
	for {
		message := <-commReceive
		convMsg := commToMsg(&message)
		assertMsg(&convMsg)
		if nw.SenderType == cl.Slave {
			if convMsg.Response != cl.Connection && convMsg.ID[0] == 'M' &&
				(convMsg.Receiver == nw.LocalIP || convMsg.Receiver == cl.All) ||
				(convMsg.Response == cl.Connection && convMsg.ID[0] == 'S') {
				nw.Receive <- convMsg
				printInfo("Slave received message", &convMsg)
			}
		}
		if nw.SenderType == cl.Master {
			if ((convMsg.ID[0] == 'S') && (convMsg.Response != cl.Connection)) ||
				((convMsg.ID[0] == 'M') && (convMsg.Response != cl.Connection) && convMsg.Receiver != nw.LocalIP) ||
				((convMsg.ID[0] == 'M') && (convMsg.Response == cl.Connection) && (convMsg.Receiver == nw.LocalIP)) {
				nw.Receive <- convMsg
				printInfo("Master received message", &convMsg)
			}
		}
	}
}

func printInfo(comment string, message *Message) {
	if ((info && message.Response != cl.Connection) || conn) && message.Response != cl.Ping {
		PrintMessage(message, comment)
	}
}

func PrintMessage(message *Message, comment string) {
	fmt.Println("__________________________________")
	fmt.Println("NETW: " + comment)
	fmt.Printf("NETW: Sender: %s\n", message.Sender)
	fmt.Printf("NETW: Receiver: %s\n", message.Receiver)
	fmt.Printf("NETW: ID: %s\n", message.ID)
	fmt.Printf("NETW: Response: %s\n", message.Response)
	fmt.Printf("NETW: Content: %v\n", message.Content)
	fmt.Println("__________________________________")
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

func Send(receiver string, senderType string, response string, content interface{}, send chan<- Message) {
	message := Message{
		Receiver: receiver,
		ID:       CreateID(senderType),
		Response: response,
		Content:  content,
	}
	send <- message
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
