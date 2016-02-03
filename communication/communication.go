package communication

import (
	// "time"
	"net"
	"fmt"
	// "bufio"
	// "os"
	"strings"
	"bytes"
)

var com_id = "2323" //Identifier for all elevators on the system
var port = ":3000"
var senderIP string


type UDPData struct {
	Identifier		string
	SenderIP			string
	ReceiverIP		string
	Data					map[string]string
}

func Init(ip string, receiveChannel chan UDPData, sendChannel chan UDPData) {
	senderIP = ip
	go listen(receiveChannel)
	go broadcast(sendChannel)
}

func udpDataToString(message *UDPData) (string){
	var buffer bytes.Buffer
	buffer.WriteString(message.Identifier)
	buffer.WriteString(" ")
	buffer.WriteString(message.SenderIP)
	buffer.WriteString(" ")
	buffer.WriteString(message.ReceiverIP)
	buffer.WriteString(" ")
	for key, value := range message.Data {
		buffer.WriteString(key)
		buffer.WriteString(":")
		buffer.WriteString(value)
		buffer.WriteString(";")
	}
	convMsg := buffer.String()
	convMsg = convMsg[:len(convMsg)-1]
	fmt.Println("Message-object converted to: " + convMsg)
	return convMsg
}

func stringToUDPData(message string) (*UDPData) {
	splitMsg := strings.Split(message, " ")
	unsplitData := strings.Split(splitMsg[3], ";")
	data := map[string]string{}
	for _, unsplitPairs := range unsplitData {
		pairs := strings.Split(unsplitPairs, ":")
		data[pairs[0]] = pairs[1]
	}
	convMsg := UDPData{
		Identifier: splitMsg[0],
		SenderIP: splitMsg[1],
		ReceiverIP: splitMsg[2],
		Data: data,
	}
	return &convMsg
}

func printError(errMsg string, err error) {
	fmt.Println(errMsg)
	fmt.Println(err)
	fmt.Println()
}

func broadcast(sendUDP chan UDPData) {
	fmt.Println("Broadcasting message to: 255.255.255.255" + port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", "255.255.255.255" + port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Broadcast failed.", err)
	}
	localAddress, err := net.ResolveUDPAddr("udp", senderIP)
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	if err != nil {
		printError("=== ERROR: DialUDP in Broadcast failed.", err)
	}
	defer connection.Close()
	for{
		message := <- sendUDP
		convMsg := udpDataToString(&message)
		connection.Write([]byte(convMsg))
		fmt.Println("Message sent successfully! \n")
	}
}

func listen(InputUDP chan UDPData) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		printError("=== ERROR: ResolvingUDPAddr in Listen failed.", err)
	}
	fmt.Print("Listening to port ")
	fmt.Println(localAddress.Port)
	connection, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		printError("=== ERROR: ListenUDP in Listen failed.", err )
	}
	defer connection.Close()
	for{
		buffer := make([]byte, 4096)
		connection.ReadFromUDP(buffer)
		message := string(buffer)
		identifier := message[:4]
		fmt.Println("Unprocessed message received: " + message)
		if (identifier == com_id) {
			convMsg := stringToUDPData(message)
			InputUDP <- *convMsg
		} else {
			fmt.Println("=== Data received ===")
			fmt.Println("Identifier does not match")
			fmt.Println(message + "\n")
		}
	}
}

func PrintMessage(data *UDPData) {
	fmt.Println("=== Data received ===")
	fmt.Println("Identifier: " + data.Identifier)
	fmt.Println("SenderIP: " + data.SenderIP)
	fmt.Println("ReceiverIP: " + data.ReceiverIP)
	fmt.Println("= Data =")
	for key, value := range data.Data {
		fmt.Println("Key: " + key + ", value: " + value)
	}
}

func Send(receiverIP string, data map[string]string, sendChannel chan UDPData) {
	message := UDPData{
		Identifier: com_id,
		SenderIP: senderIP,
		ReceiverIP: receiverIP,
		Data: data,
	}
	sendChannel <- message
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