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
		backup.Run("")
	}
	go slave.Run()
}


func remoteInstall() {
	elevIP := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("This is a simple script for installing elevators on remote computers. \n Used on own responsibility, as this script is VERY simple and won't like input that contains any errors")
	fmt.Print("Start remote elevators? (y/n) ")
	ans, _ := reader.ReadString('\n')
	if ans == "y" || ans == "Y" {
		for ans == "y" || ans == "Y" {
			fmt.Print("Insert a IP-address: ")
			ans, _ = reader.ReadString('\n')
			elevIP = append(elevIP, ans)
			fmt.Print("Add another elevator? (y/n) ")
			ans, _ = reader.ReadString('\n')
		}
		fmt.Println("Installing...")
		for elev := range elveIP {
			command1 := "ssh student@" + elev + ";"
			command2 := "mkdir $HOME/Desktop/gospace;"
			command3 := "export GOPATH=$HOME/Desktop/gospace;"
			command4 := "go get github.com/holwech/heislab;"
			command5 := "go get github.com/satori/go.uuid;"
			command6 := "go run $GOPATH/src/github.com/holwech/heislab/main.go;"
			cmd = exec.Command(command1, command2, command3, command4, command5, command6)
			cmd.Run()
		}
	}
}
