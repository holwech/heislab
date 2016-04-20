package backup

import (
	"testing"
	"fmt"
)

func TestBackup(t *testing.T) {
	backup("-f")
	for{

	}
}

func TestListen(t *testing.T) {
	timeout := listen()
	<- timeout
	fmt.Println("Timed out")
}
