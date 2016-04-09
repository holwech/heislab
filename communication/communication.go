package communication

import (
	"net"
	"fmt"
	"os"
	"encoding/json"
	"time"
)

const com_id = "2323" //Identifier for all elevators on the system
const port = ":22212"
const broadcast_addr = "255.255.255.255"
type InnerOrder struct{
	Floor int `json:"Floor"`
}
type OuterOrder struct{
	Floor int `json:"Floor"`
	Direction int `json:"Direction"`
}
// DataValue should ONLY be int og string
type CommData struct {
	Identifier string
	SenderIP	string
	ReceiverIP	string
	MsgID string
	DataType string
	DataValue interface{}
}

type ConnData struct {
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

func Run(sendCh chan CommData) (<- chan CommData, <- chan ConnData) {
	commReceive := make(chan CommData)
	commSentStatus := make(chan ConnData)
	commSend := make(chan CommData)
	connStatus := make(chan ConnData)
	receivedMsg := make(chan CommData)
	go listen(commReceive)
	go broadcast(commSend, commSentStatus)
	go checkTimeout(commSentStatus, connStatus)
	go msgSorter(commReceive, receivedMsg, commSentStatus, commSend, sendCh)
	return receivedMsg, connStatus
}

func msgSorter(commReceive <-chan CommData, receivedMsg chan<- CommData, commSentStatus chan<- ConnData, commSend chan<- CommData, sendCh <-chan CommData) {
	for{
		select{
		// When messages are received
		case message := <- commReceive:
			// If message is a receive-confirmation, push to status-channel
			if message.DataType == "OK"{
				// Filters out status-messages that are not relevant for receiver
				if message.SenderIP == GetLocalIP() {
					response := ConnData{
						SenderIP: message.SenderIP,
						MsgID: message.MsgID,
						SendTime: time.Now(),
						Status: "OK",
					}
					commSentStatus <- response
				}
			// If message is a normal message, then send verification 
			}else{
				response := CommData{
					Identifier: com_id,
					SenderIP: message.SenderIP, 
					ReceiverIP: GetLocalIP(),
					MsgID: message.MsgID,
					DataType: "OK",
					DataValue: time.Now(),
				}
				receivedMsg <- message
				commSend <- response
			}
		// When messages are sent, set time-stamp
		case message := <- sendCh:
			commSend <- message
			timeSent := ConnData{
				SenderIP: message.SenderIP,
				MsgID: message.MsgID,
				SendTime: time.Now(),
				Status: "Sent",
			}
			commSentStatus <- timeSent
		}
	}
}

func checkTimeout(commSentStatus chan ConnData, connStatus chan ConnData) {
	messageLog := make(map[string]ConnData)
	ticker := time.NewTicker(50 * time.Millisecond).C
	for{
		select{
		case metadata := <- commSentStatus:
			if metadata.Status == "OK" {
				delete(messageLog, metadata.MsgID)
				fmt.Printf("COMM: Message received, sending verification. ID: %s\n", metadata.MsgID)
				connStatus <- metadata
			}else{
				messageLog[metadata.MsgID] = metadata
				fmt.Printf("COMM: Metadata stored\n")
			}
		case <- ticker:
			currentTime := time.Now()
			for msgID, metadata := range messageLog {
				timeDiff := currentTime.Sub(metadata.SendTime)
				if timeDiff.Seconds() > 5 {
					sendingFailed := metadata
					sendingFailed.Status = "Failed"
					delete(messageLog, msgID)
					connStatus <- sendingFailed
				}
			}
		}
	}
}



func broadcast(sendCh chan CommData, commSentStatus chan ConnData) {
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
		message := <- sendCh
		convMsg, err := json.Marshal(message)
		if err != nil {
			printError("=== ERROR: Convertion of json failed in broadcast", err)
		}
		connection.Write(convMsg)
		fmt.Printf("COMM: Message sent successfully!\n")
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
		if (message.Identifier == com_id) {
			fmt.Printf("COMM: Message received from: %s\n", message.SenderIP)
			commReceive <- message
		} else {
			fmt.Printf("COMM: Data received\n")
			fmt.Printf("COMM: Identifier does not match\n")
			fmt.Printf("COMM: %s\n\n", string(buffer))
		}
	}
}

func PrintMessage(data CommData) {
	fmt.Printf("=== Data received ===\n")
	fmt.Printf("Identifier: %s\n", data.Identifier)
	fmt.Printf("SenderIP: %s\n", data.SenderIP)
	fmt.Printf("ReceiverIP: %s\n", data.ReceiverIP)
	fmt.Printf("Message ID: %s\n", data.MsgID)
	fmt.Printf("= Data = \n")
	fmt.Printf("Data type: %s\n", data.DataType)
	fmt.Printf("DataValue: %s\n", data.DataValue)
}

func PrintConnData(data ConnData) {
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

func ResolveMsg(receiverIP string, msgID string, dataType string, dataValue interface{}) (commData *CommData) {
	message := CommData{
		Identifier: com_id,
		SenderIP: GetLocalIP(),
		ReceiverIP: receiverIP,
		MsgID: msgID,
		DataType: dataType,
		DataValue: dataValue,
	}
	return &message
}

// func SendConsoleMsg(config *config, sendUDP chan UDPData) {
// 	time.Sleep(1*time.Second)
// 	fmt.Println("=== Send from console ===")
// 	terminate := "y\n"
// 	for terminate == "y\n" {
// 		reader := bufio.NewReader(os.Stdin)
// 		message := &UDPData{
// 			Identifier: com_id,
// 			SenderIP: config.SenderIP,
// 			ReceiverIP: config.ReceiverIP,
// 			Data: map[string]string{},
// 		}
// 		for terminate == "y\n" {
// 			fmt.Print("Enter key: ")
// 			key, _ := reader.ReadString('\n')
// 			fmt.Print("Enter value: ")
// 			value, _ := reader.ReadString('\n')
// 			message.Data[key] = value
// 			fmt.Print("Add more data values? (y/n): ")
// 			terminate, _ = reader.ReadString('\n')
// 			fmt.Println(terminate)
// 		}
// 		sendUDP <- *message
// 		time.Sleep(1*time.Second)
// 		fmt.Print("Send another message? (y/n): ")
// 		terminate, _ = reader.ReadString('\n')
// 	}
// 	fmt.Println("=== Stopping send from console ===")
// }
