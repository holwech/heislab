package communication

import (
	"testing"
	"time"
	"fmt"
)

func TestMain(t *testing.T) {
	sendCh := make(chan CommData)
	receiveCh,_ := Run(sendCh)
	time.Sleep(1 * time.Second)
	count := 0
	for{
		msg := ResolveMsg("10.20.78.108", "Test", "Penis")
		sendCh <- *msg
		time.Sleep(10 * time.Second)
		PrintMessage(<-receiveCh)
		fmt.Println(count)
		count += 1
	}
}
