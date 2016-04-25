package communication

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/holwech/heislab/cl"
)

const com_id = "2323" //Key for all elevators on the system
const broadcast_addr = "255.255.255.255"

type Communication struct {
	CommReceive, CommSend, Receive, Send chan CommData
	LocalIP, ReadPort, WritePort string
}

type CommData struct {
	Key, SenderIP, ReceiverIP, MsgID, Response string
	Content interface{}
}

type Timestamp struct {
	SenderIP, MsgID, Status string
	SendTime time.Time
}

func printError(errMsg string, err error) {
	if err != nil {
		fmt.Printf(errMsg + "\n")
		fmt.Printf(err.Error() + "\n")
		fmt.Println()
	}
}

func (cm *Communication) Init(readPort, writePort string) {
	cm.CommReceive = make(chan CommData, 10)
	cm.CommSend = make(chan CommData)
	cm.Receive = make(chan CommData)
	cm.Send = make(chan CommData)
	cm.LocalIP = GetLocalIP()
	cm.ReadPort = readPort
	cm.WritePort = writePort
}

func Init(readPort string, writePort string) (<-chan CommData, chan<- CommData) {
	cm := new(Communication)
	cm.Init(readPort, writePort)
	run(cm)
	return cm.Receive, cm.Send
}

func run(cm *Communication) 
	go listen(cm.CommReceive, cm.ReadPort)
	go broadcast(cm.CommSend, cm.LocalIP, cm.WritePort)
	go msgSorter(cm)
}

func msgSorter(cm *Communication) {
	messageLog := make(map[string]Timestamp)
	ticker := time.NewTicker(50 * time.Millisecond).C
	for {
		select {
		case message := <- cm.CommReceive:
			// If message is a receive-confirmation, push to status-channel
			if message.Response == cl.Connection {
				// Filters out status-messages that are not relevant for receiver
				_, exists := messageLog[message.MsgID]
				if message.ReceiverIP == cm.LocalIP && message.Content == cl.OK && exists{
					delete(messageLog, message.MsgID)
					status := CommData{
						Key:        com_id,
						SenderIP:   message.SenderIP,
						ReceiverIP: cm.LocalIP,
						MsgID:      message.MsgID,
						Response:   cl.Connection,
						Content:    cl.OK,
					}
					cm.Receive <- status
				}
				// If message is a normal message, then send verification
			} else {
				cm.Receive <- message
				if message.ReceiverIP == cm.LocalIP {
					ok := CommData{
						Key:        com_id,
						SenderIP:   cm.LocalIP,
						ReceiverIP: message.SenderIP,
						MsgID:      message.MsgID,
						Response:   cl.Connection,
						Content:    cl.OK,
					}
					cm.CommSend <- ok
				}
			}
		// When messages are sent, set time-stamp
		case message := <- cm.Send:
			timeSent := Timestamp{
				SenderIP: message.SenderIP,
				MsgID:    message.MsgID,
				SendTime: time.Now(),
				Status:   cl.Sent,
			}
			cm.CommSend <- message
			messageLog[message.MsgID] = timeSent
		case <-ticker:
			currentTime := time.Now()
			for msgID, metadata := range messageLog {
				timeDiff := currentTime.Sub(metadata.SendTime)
				if timeDiff > 50 * time.Millisecond {
					delete(messageLog, msgID)
					status := CommData{
						Key:        com_id,
						SenderIP:   metadata.SenderIP,
						ReceiverIP: cm.LocalIP,
						MsgID:      metadata.MsgID,
						Response:   cl.Connection,
						Content:    cl.Failed,
					}
					cm.Receive <- status
				}
			}
		}
	}
}


func broadcast(commSend chan CommData, localIP string, port string) {
	fmt.Printf("COMM: Broadcasting message to: %s%s\n", broadcast_addr, port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", broadcast_addr+port)
	printError("=== ERROR: ResolvingUDPAddr in Broadcast failed.", err)
	localAddress, err := net.ResolveUDPAddr("udp", GetLocalIP())
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	printError("=== ERROR: DialUDP in Broadcast failed.", err)
	defer connection.Close()
	for {
		message := <-commSend
		convMsg, err := json.Marshal(message)
		printError("=== ERROR: Convertion of json failed in broadcast", err)
		_, err = connection.Write(convMsg)
		printError("=== ERROR: Write in broadcast failed", err)
	}
}

func listen(commReceive chan CommData, port string) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	printError("=== ERROR: ResolvingUDPAddr in Listen failed.", err)
	fmt.Printf("COMM: Listening to port %d\n", localAddress.Port)
	connection, err := net.ListenUDP("udp", localAddress)
	printError("=== ERROR: ListenUDP in Listen failed.", err)
	defer connection.Close()
	for {
		var message CommData
		buffer := make([]byte, 4096)
		length, _, err := connection.ReadFromUDP(buffer)
		printError("=== ERROR: ReadFromUDP failed in listen", err)
		buffer = buffer[:length]
		err = json.Unmarshal(buffer, &message)
		printError("=== ERROR: Unmarshal failed in listen", err)
		//Filters out all messages not relevant for the system
		if message.Key == com_id {
			commReceive <- message
		}
	}
}

func PrintMessage(data CommData) {
	fmt.Printf("=== Data received ===\n")
	fmt.Printf("Key: %s\n", data.Key)
	fmt.Printf("SenderIP: %s\n", data.SenderIP)
	fmt.Printf("ReceiverIP: %s\n", data.ReceiverIP)
	fmt.Printf("Message ID: %s\n", data.MsgID)
	fmt.Printf("= Data = \n")
	fmt.Printf("Data type: %s\n", data.Response)
	fmt.Printf("Content: %v\n", data.Content)
}

func PrintTimestamp(data Timestamp) {
	fmt.Printf("=== Connection data ===\n")
	fmt.Printf("SenderIP: %s\n", data.SenderIP)
	fmt.Printf("Message ID: %s\n", data.MsgID)
	fmt.Printf("Time: %s\n", data.SendTime)
	fmt.Printf("Status: %s\n", data.Status)
}

func GetLocalIP() string {
	var localIP string
	addr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("=== ERROR: GetLocalIP in communication failed")
		os.Exit(1)
	}
	for _, val := range addr {
		if ip, ok := val.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				localIP = ip.IP.String()
			}
		}
	}
	return localIP
}

func ResolveMsg(senderIP string, receiverIP string, msgID string, response string, content interface{}) (commData *CommData) {
	message := CommData{
		Key:        com_id,
		SenderIP:   senderIP,
		ReceiverIP: receiverIP,
		MsgID:      msgID,
		Response:   response,
		Content:    content,
	}
	return &message
}
