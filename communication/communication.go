package communication

import (
	"net"
	"fmt"
	"os"
	"encoding/json"
	"time"
	"github.com/fatih/color"
)

const com_id = "2323" //Identifier for all elevators on the system
const port = ":3000"

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
	color.Blue(errMsg + "\n")
	color.Blue(err.Error() + "\n")
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
			// If message is a receive-confirmation
			if message.DataType == "Received"{
				if message.SenderIP == GetLocalIP(){
					response := ConnData{
						SenderIP: message.SenderIP,
						MsgID: message.MsgID,
						SendTime: time.Now(),
						Status: "Received",
					}
					commSentStatus <- response
				}
				// If message is a normal message 
			}else{
				response := CommData{
					Identifier: com_id,
					SenderIP: GetLocalIP(),
					ReceiverIP: message.SenderIP,
					MsgID: message.MsgID,
					DataType: "Received",
					DataValue: time.Now(),
				}
				receivedMsg <- message
				commSend <- response
			}
			// When messages are sent
		case message := <- sendCh:
			commSend <- message
			timeSent := ConnData{
				SenderIP: GetLocalIP(),
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
			if metadata.Status == "Received" {
				delete(messageLog, metadata.MsgID)
				color.Blue("COMM: Message received, sending verification. ID: %s\n", metadata.MsgID)
				connStatus <- metadata
			}else{
				messageLog[metadata.MsgID] = metadata
				color.Blue("COMM: Metadata stored\n")
			}
		case <- ticker:
			currentTime := time.Now()
			for msgID, metadata := range messageLog {
				timeDiff := currentTime.Sub(metadata.SendTime)
				if timeDiff.Seconds() > 0.50 {
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
	color.Blue("COMM: Broadcasting message to: 255.255.255.255%s\n", port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", "255.255.255.255" + port)
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
		color.Blue("COMM: Message sent successfully!\n")
	}
}

func listen(commReceive chan CommData) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Listen failed.", err)
	}
	color.Blue("COMM: Listening to port %d\n", localAddress.Port)
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
		if (message.Identifier == com_id) {
			color.Blue("COMM: Message received from: ")
			color.Blue("%s\n", message.SenderIP)
			commReceive <- message
		} else {
			color.Blue("COMM: Data received\n")
			color.Blue("COMM: Identifier does not match\n")
			color.Blue("COMM: %s\n\n", string(buffer))
		}
	}
}

func PrintMessage(data CommData) {
	color.Blue("=== Data received ===\n")
	color.Blue("Identifier: %s\n", data.Identifier)
	color.Blue("SenderIP: %s\n", data.SenderIP)
	color.Blue("ReceiverIP: %s\n", data.ReceiverIP)
	color.Blue("Message ID: %s\n", data.MsgID)
	color.Blue("= Data = \n")
	color.Blue("Data type: %s\n", data.DataType)
	color.Blue("DataValue: %s\n", data.DataValue)
}

func PrintConnData(data ConnData) {
	color.Blue("=== Connection data ===\n")
	color.Blue("SenderIP: %s\n", data.SenderIP)
	color.Blue("Message ID: %s\n", data.MsgID)
	color.Blue("Time: %s\n", data.SendTime)
	color.Blue("Status: %s\n", data.Status)
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
