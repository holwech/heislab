package main

import (
	"fmt"
	"time"
	"github.com/holwech/heislab/communication"
)

func main() {
	test()
}



func test() {
	communicationData := make(chan communication.CommData)
	sendCh := make(chan communication.CommData)
	go communication.Run(communicationData, sendCh)
	time.Sleep(1 * time.Second)
	count := 0
	for{
		communication.Send("10.20.78.108", "Test", "Penis", sendCh)
		time.Sleep(10 * time.Second)
		communication.PrintMessage(<- communicationData)
		fmt.Println(count)
		count += 1
	}
}
