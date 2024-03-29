package main

import (
	"fmt"
	"github.com/holwech/heislab/backup"
	"github.com/holwech/heislab/slave"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Remove?
	fmt.Println("Starting elevator")
	if len(os.Args) > 1 {
		backup.Run(os.Args[1])
		slave.Run(true)
	} else {
		backup.Run("")
		slave.Run(false)
	}
}
