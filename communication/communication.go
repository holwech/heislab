package communication

import (
	"encoding/json"
	"fmt"
	"github.com/holwech/heislab/cl"
	"net"
)

const com_id = "2323" //Key for all elevators in the system
const broadcast_addr = "255.255.255.255"

type CommData struct {
	Key, SenderIP, ReceiverIP, MsgID, Response string
	Content                                    interface{}
}

func Init(readPort string, writePort string) (<-chan CommData, chan<- CommData) {
	receive := make(chan CommData, 10*cl.Floors)
	send := make(chan CommData, 10*cl.Floors)
	localIP := GetLocalIP()
	go listen(receive, readPort)
	go broadcast(send, localIP, writePort)
	return receive, send
}

func broadcast(send chan CommData, localIP string, port string) {
	fmt.Printf("COMM: Broadcasting message to: %s%s\n", broadcast_addr, port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", broadcast_addr+port)
	printError("ResolvingUDPAddr in Broadcast failed.", err)
	localAddress, err := net.ResolveUDPAddr("udp", GetLocalIP())
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	printError("DialUDP in Broadcast failed.", err)

	localhostAddress, err := net.ResolveUDPAddr("udp", "localhost"+port)
	printError("ResolvingUDPAddr in Broadcast localhost failed.", err)
	lConnection, err := net.DialUDP("udp", localAddress, localhostAddress)
	printError("DialUDP in Broadcast localhost failed.", err)
	defer connection.Close()
	for {
		message := <-send
		convMsg, err := json.Marshal(message)
		printError("Convertion of json failed in broadcast", err)
		_, err = connection.Write(convMsg)
		if err != nil {
			_, err = lConnection.Write(convMsg)
			printError("Write in broadcast localhost failed", err)
		}
	}
}

func listen(receive chan CommData, port string) {
	localAddress, err := net.ResolveUDPAddr("udp", port)
	printError("ResolvingUDPAddr in Listen failed.", err)
	fmt.Printf("COMM: Listening to port %d\n", localAddress.Port)
	connection, err := net.ListenUDP("udp", localAddress)
	printError("ListenUDP in Listen failed.", err)
	defer connection.Close()
	for {
		var message CommData
		buffer := make([]byte, 4096)
		length, _, err := connection.ReadFromUDP(buffer)
		printError("ReadFromUDP failed in listen", err)
		buffer = buffer[:length]
		err = json.Unmarshal(buffer, &message)
		printError("Unmarshal failed in listen", err)
		//Filters out all messages not relevant for the system
		if message.Key == com_id {
			receive <- message
		}
		if message.Response == cl.Backup {
			PrintMessage(message)
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

func printError(errMsg string, err error) {
	if err != nil {
		fmt.Println(errMsg)
		fmt.Println(err.Error())
	}
}

func GetLocalIP() string {
	var localIP string
	addr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("GetLocalIP in communication failed")
		return "localhost"
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
