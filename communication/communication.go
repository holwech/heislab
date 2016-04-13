package communication

import (
	"net"
	"fmt"
	"os"
	"encoding/json"
	"time"
	"github.com/holwech/heislab/cl"
)

const com_id = "2323" //Key for all elevators on the system
const port = ":22212"
const broadcast_addr = "255.255.255.255"

// Content should ONLY be int og string
type CommData struct {
	Key string
	SenderIP	string
	ReceiverIP	string
	MsgID string
	Response string
	Content interface{}
}

type Timestamp struct {
	SenderIP string
	MsgID string
	SendTime time.Time
	Status string
}

func printError(errMsg string, err error) {
	fmt.Printf(errMsg + "\n")
	fmt.Printf(err.Error() + "\n")
	fmt.Println()
}

func Run(sendCh chan CommData) (<- chan CommData) {
	commReceive := make(chan CommData)
	connStatus := make(chan Timestamp)
	commSend := make(chan CommData)
	receivedMsg := make(chan CommData)
	go listen(commReceive)
	go broadcast(commSend, connStatus)
	go checkTimeout(connStatus, receivedMsg)
	go msgSorter(commReceive, receivedMsg, connStatus, commSend, sendCh)
	return receivedMsg
}

func msgSorter(commReceive <-chan CommData, receivedMsg chan<- CommData, connStatus chan<- Timestamp, commSend chan<- CommData, sendCh <-chan CommData) {
	for{
		select{
		// When messages are received
		case message := <- commReceive:
			// If message is a receive-confirmation, push to status-channel
			if message.Response == cl.Connection{
				// Filters out status-messages that are not relevant for receiver
				if message.ReceiverIP == GetLocalIP() {
					received := Timestamp{
						SenderIP: message.SenderIP,
						MsgID: message.MsgID,
						SendTime: time.Now(),
						Status: cl.OK,
					}
					connStatus <- received
				}
			// If message is a normal message, then send verification 
			}else{
				ok := CommData{
					Key: com_id,
					SenderIP: message.ReceiverIP,
					ReceiverIP: message.SenderIP,
					MsgID: message.MsgID,
					Response: cl.Connection,
					Content: cl.OK,
				}
				receivedMsg <- message
				commSend <- ok
			}
		// When messages are sent, set time-stamp
		case message := <- sendCh:
			commSend <- message
			timeSent := Timestamp{
				SenderIP: message.SenderIP,
				MsgID: message.MsgID,
				SendTime: time.Now(),
				Status: cl.Sent,
			}
			connStatus <- timeSent
		}
	}
}

func checkTimeout(connStatus chan Timestamp, receivedMsg chan CommData) {
	messageLog := make(map[string]Timestamp)
	ticker := time.NewTicker(50 * time.Millisecond).C
	for{
		select{
		case metadata := <- connStatus:
			if metadata.Status == cl.OK {
				delete(messageLog, metadata.MsgID)
				status := CommData{
					Key: com_id,
					SenderIP: metadata.SenderIP,
					ReceiverIP: GetLocalIP(),
					MsgID: metadata.MsgID,
					Response: cl.Connection,
					Content: cl.OK,
				}
				receivedMsg <- status
			}else{
				messageLog[metadata.MsgID] = metadata
			}
		case <- ticker:
			currentTime := time.Now()
			for msgID, metadata := range messageLog {
				timeDiff := currentTime.Sub(metadata.SendTime)
				if timeDiff.Seconds() > 5 {
					delete(messageLog, msgID)
					status := CommData{
						Key: com_id,
						SenderIP: metadata.SenderIP,
						ReceiverIP: GetLocalIP(),
						MsgID: metadata.MsgID,
						Response: cl.Connection,
						Content: cl.Failed,
					}
					receivedMsg <- status
				}
			}
		}
	}
}



func broadcast(commSend chan CommData, connStatus chan Timestamp) {
	fmt.Printf("COMM: Broadcasting message to: %s%s\n", broadcast_addr, port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", broadcast_addr + port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Broadcast failed.", err)
	}
	localAddress, err := net.ResolveUDPAddr("udp", GetLocalIP())
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	if err != nil {
		printError("=== ERROR: DialUDP in Broadcast failed.", err)
	}
	defer connection.Close()
	for{
		message := <- commSend
		convMsg, err := json.Marshal(message)
		if err != nil {
			printError("=== ERROR: Convertion of json failed in broadcast", err)
		}
		connection.Write(convMsg)
	}
}

func listen(commReceive chan CommData) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Listen failed.", err)
	}
	fmt.Printf("COMM: Listening to port %d\n", localAddress.Port)
	connection, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		printError("=== ERROR: ListenUDP in Listen failed.", err )
	}
	defer connection.Close()
	for{
		var message CommData
		buffer := make([]byte, 4096)
		length, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			printError("=== ERROR: ReadFromUDP failed in listen", err)
		}
		buffer = buffer[:length]
		err = json.Unmarshal(buffer, &message)
		if err != nil {
			printError("=== ERROR: Unmarshal failed in listen", err)
		}
		//Filters out all messages not relevant for the system
		if (message.Key == com_id) {
			commReceive <- message
		} else {
			fmt.Printf("COMM: Data received\n")
			fmt.Printf("COMM: Key does not match\n")
			fmt.Printf("COMM: %s\n\n", string(buffer))
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

func GetLocalIP() (string) {
	var localIP string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP = ipnet.IP.String()
			}
		}
	}
	return localIP
}

func ResolveMsg(senderIP string, receiverIP string, msgID string, response string, content interface{}) (commData *CommData) {
	message := CommData{
		Key: com_id,
		SenderIP: senderIP,
		ReceiverIP: receiverIP,
		MsgID: msgID,
		Response: response,
		Content: content,
	}
	return &message
}
