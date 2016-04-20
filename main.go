package main

import (
	"github.com/holwech/heislab/backup"
	"os"
	"fmt"
)

func main() {
	fmt.Println("Starting elevator")
	if len(os.Args) > 1 {
		backup.Run(os.Args[1])
	} else {
		backup.Run("-f")
	}
	for{}
}
