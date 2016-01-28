package main

import (
	"github.com/holwech/heislab/communication"
	"fmt"
)

func main() {
	communicationTest()
}


func communicationTest() {
	InputUDP := make (chan string)
	go communication.Broadcast()
	go communication.Listen(InputUDP)
	for{
		fmt.Println(<- InputUDP)
	}
}