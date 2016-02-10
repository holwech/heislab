package communication

import (
	"net"
	"fmt"
	"encoding/json"
)

var com_id = "2323" //Identifier for all elevators on the system
var port = ":3000"
var senderIP string


type UDPData struct {
	Identifier		string
	SenderIP			string
	ReceiverIP		string
	Data					map[string]interface{}
}

func Init(ip string, receiveChannel chan UDPData, sendChannel chan UDPData) {
	senderIP = ip
	go listen(receiveChannel)
	go broadcast(sendChannel)
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
		//convMsg := udpDataToString(&message)
		convMsg, err := json.Marshal(message)
		if err != nil {
			printError("=== ERROR: Convertion of json failed in broadcast", err)
		}
		connection.Write(convMsg)
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
		var message UDPData
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
		fmt.Println("Message received from: ")
		fmt.Println(message.SenderIP)
		if (message.Identifier == com_id) {
			InputUDP <- message
		} else {
			fmt.Println("=== Data received ===")
			fmt.Println("Identifier does not match")
			fmt.Println(string(buffer) + "\n")
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
		fmt.Print("Key: ")
		fmt.Print(key)
		fmt.Print(", value: ")
		fmt.Println(value)
	}
}

func Send(receiverIP string, data map[string]interface{}, sendChannel chan UDPData) {
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