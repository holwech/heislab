dpackage main

import (
	"github.com/holwech/heislab/communication"
	"time"
)

func main() {
	communicationTest()
}


func communicationTest() {
	// receiveChannel := make(chan communication.UDPData)
	// sendChannel := make(chan communication.UDPData)
	// config := &communication.Config{
	// 	SenderIP: "192.168.1.3",
	// 	ReceiverIP: "255.255.255.255",
	// 	Port: ":30000",
	// }
	// message1 := &communication.UDPData{
	// 	Identifier: "2323",
	// 	SenderIP: "192.168.1.3",
	// 	ReceiverIP: "255.255.255.255",
	// 	Data: map[string]string{
	// 		"Command": "UP",
	// 		"Door": "OPEN",
	// 	},
	// }
	// message2 := &communication.UDPData{
	// 	Identifier: "2329292923",
	// 	SenderIP: "192.168.1.3",
	// 	ReceiverIP: "255.255.255.255",
	// 	Data: map[string]string{
	// 		"Command": "UP",
	// 		"Door": "OPEN",
	// 	},
	// }
	// go communication.Listen(config, receiveChannel)
	// go communication.Broadcast(config, sendChannel)
	// time.Sleep(1*time.Second)

	// sendChannel <- *message1
	// sendChannel <- *message2
	// go communication.SendConsoleMsg(config, sendChannel)
	// for{
	// 	data := <- receiveChannel
	// 	communication.PrintMessage(&data)
	// }
	data := map[string]interface{} {
		"LOL": 1,
		"FAKA U BTCH": "U EAT MY NUDLS",
	}
	receiveChannel := make(chan communication.UDPData)
	sendChannel := make(chan communication.UDPData)
	communication.Init("10.20.78.108", receiveChannel, sendChannel)
	time.Sleep(1*time.Second)
	communication.Send("10.20.78.108", data, sendChannel)
	for {
		message := <- receiveChannel
		communication.PrintMessage(&message)
	}
}