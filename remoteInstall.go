package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func remoteInstall() {
	elevIP := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("This is a simple script for installing elevators on remote computers. \n Used on own responsibility, as this script is VERY simple and won't like input that contains any errors \n Good luck!")
	fmt.Print("Start remote elevators? (y/n) ")
	ans, _ := reader.ReadString('\n')
	if ans == "y\n" || ans == "Y\n" {
		for ans == "y\n" || ans == "Y\n" {
			fmt.Print("Insert a IP-address: ")
			ans, _ = reader.ReadString('\n')
			elevIP = append(elevIP, ans)
			fmt.Print("Add another elevator? (y/n) ")
			ans, _ = reader.ReadString('\n')
		}
		fmt.Println("Installing...")
		for _, elev := range elevIP {
			command1 := "ssh student@" + elev + ";"
			command2 := "mkdir $HOME/Desktop/gospace;"
			command3 := "export GOPATH=$HOME/Desktop/gospace;"
			command4 := "go get github.com/holwech/heislab;"
			command5 := "go get github.com/satori/go.uuid;"
			command6 := "go run $GOPATH/src/github.com/holwech/heislab/main.go;"
			cmd := exec.Command(command1, command2, command3, command4, command5, command6)
			cmd.Run()
			fmt.Println("Elevator " + elev + " installed and started (maybe)")
		}
		fmt.Println("Success! (again, maybe)")
	}
	fmt.Println("Exiting remote install")
}
func main() {
	remoteInstall()
}
