package communication

import (
	"net"
	"fmt"
	"os"
	"encoding/json"
)

const com_id = "2323" //Identifier for all elevators on the system
const port = ":3000"


type CommData struct {
	Identifier string
	SenderIP	string
	ReceiverIP	string
	MsgID: int32
	DataType string
	DataValue interface{}
}

type ConnData struct {
	SenderIP string
	MsgID int32
	SendTime: time.Time
	Status string
}

func printError(errMsg string, err error) {
	fmt.Println(errMsg)
	fmt.Println(err)
	fmt.Println()
}

func Run(receivedMsg chan CommData, sendCh chan CommData, connStatus chan ConnData) {
	receiveCh := make(chan CommData)
	metadata := make(chan ConnData)
	sendingStatus := make(chan ConnData)
	go listen(receiveCh)
	go broadcast(sendCh, metadata)
	go checkTimeout(timeSent, sendingStatus)
	for{
		select{
			case message := <- receiveCh:
				if message.
			case message := <- sendingStatus
				connStatus <- message
		}
	}
}


func checkTimeout(timeSent chan ConnData, sendingStatus chan ConnData) {
	messageLog := map[int32][ConnData]
	ticker := time.NewTicker(50 * time.Milliseconds)
	for{
		select{
		case metadata := <- timeSent:
			messageLog[metadata.MsgID] = metadata
		case <- ticker:
			currentTime := time.Now()
			for msgID, metadata := range messageLog {
				if currentTime.Sub(metadata.sendTime) > 0.050 {
					sendingFailed := metadata
					sendingFailed.Status = "Failed"
					delete(messageLog, msgID)
					sendingStatus <- sendingFailed
				}
			}
		}
	}
}



func broadcast(sendCh chan CommData, metadata chan ConnData) {
	fmt.Println("COMM: Broadcasting message to: 255.255.255.255" + port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", "255.255.255.255" + port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Broadcast failed.", err)
	}
	localAddress, err := net.ResolveUDPAddr("udp", getLocalIP())
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	if err != nil {
		printError("=== ERROR: DialUDP in Broadcast failed.", err)
	}
	defer connection.Close()
	timeSent := ConnData{}
	msgID int32 = 0
	var sendTime time.Time
	for{
		message := <- sendCh
		message.MsgID = msgID
		convMsg, err := json.Marshal(message)
		if err != nil {
			printError("=== ERROR: Convertion of json failed in broadcast", err)
		}
		connection.Write(convMsg)
		fmt.Println("COMM: Message sent successfully! \n")

		sendTime = time.Now()
		timeSent{
			"SenderIP": localAddress.IP,
			"MsgID": msgID,
			"SendTime": time.Now(),
			"Status": "Sent",
		}
		msgID += 1
	}
}

func listen(receivedMsg chan CommData) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Listen failed.", err)
	}
	fmt.Print("COMM: Listening to port ")
	fmt.Println(localAddress.Port)
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
		fmt.Print("COMM: Message received from: ")
		fmt.Println(message.SenderIP)
		if (message.Identifier == com_id) {
			receivedMsg <- message
		} else {
			fmt.Println("COMM: Data received")
			fmt.Println("COMM: Identifier does not match")
			fmt.Println("COMM: " + string(buffer) + "\n")
		}
	}
}

func PrintMessage(data CommData) {
	fmt.Println("=== Data received ===")
	fmt.Println("Identifier: " + data.Identifier)
	fmt.Println("SenderIP: " + data.SenderIP)
	fmt.Println("ReceiverIP: " + data.ReceiverIP)
	fmt.Println("= Data =")
	fmt.Println("Data type: " + data.DataType)
	fmt.Print("DataValue: ")
	fmt.Println(data.DataValue)
}

func getLocalIP() (string) {
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

func Send(receiverIP string, dataType string, dataValue interface{}, sendCh chan CommData) {
	message := CommData{
		Identifier: com_id,
		SenderIP: getLocalIP(),
		ReceiverIP: receiverIP,
		MsgID: 0,
		DataType: dataType,
		DataValue: dataValue,
	}
	sendCh <- message
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
