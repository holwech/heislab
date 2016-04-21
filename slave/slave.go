package slave

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
	"time"
	"os/exec"
	"bufio"
	"os"
)

type Slave struct {
	DoorTimer, StartupTimer, MotorTimer *time.Timer
	MasterID, EngineState               string
}

func (sl *Slave) Init() {
	sl.DoorTimer = time.NewTimer(time.Second)
	sl.DoorTimer.Stop()
	sl.StartupTimer = time.NewTimer(time.Second)
	sl.StartupTimer.Stop()
	sl.MotorTimer = time.NewTimer(time.Second)
	sl.MotorTimer.Stop()
	sl.EngineState = cl.EngineOK
	sl.MasterID = cl.Unknown
}

func initSlave() *Slave {
	sl := new(Slave)
	sl.Init()
	return sl
}

func Run() {
	innerChan, outerChan, floorChan := driver.InitElevator()
	nw := network.InitNetwork()
	master.InitMaster(nw)
	sl := initSlave()
	slaveReceive, slaveSend := nw.SChannels()
	send(nw.LocalIP, "", cl.System, cl.Startup, slaveSend)

	ticker := time.NewTicker(3*time.Second)
	sl.StartupTimer.Reset(50 * time.Millisecond)
	go remoteInstall()
	for {
		select {
		case innerOrder := <-innerChan:
			send(sl.MasterID, "", cl.InnerOrder, innerOrder, slaveSend)
		case outerOrder := <-outerChan:
			send(sl.MasterID, "", cl.OuterOrder, outerOrder, slaveSend)
		case newFloor := <-floorChan:
			fmt.Printf("Floor: %d\n", newFloor)
			send(sl.MasterID, "", cl.Floor, newFloor, slaveSend)
			if newFloor != -1 {
				sl.MotorTimer.Reset(6 * time.Second)
				if sl.EngineState == cl.EngineFail {
					sl.EngineState = cl.EngineOK
					send(sl.MasterID, "", cl.System, cl.EngineOK, slaveSend)
				}
			}
		case <-sl.DoorTimer.C:
			driver.SetDoorLamp(0)
			send(sl.MasterID, "", cl.DoorClosed, "", slaveSend)
		case message := <-slaveReceive:
			handleInput(sl, nw, message, slaveSend)
		case <-sl.StartupTimer.C:
			send(nw.LocalIP, "", cl.System, cl.SetMaster, slaveSend)
			sl.MasterID = nw.LocalIP
		case <-sl.MotorTimer.C:
			sl.EngineState = cl.EngineFail
			send(sl.MasterID, "", cl.System, cl.EngineFail, slaveSend)
		case <- ticker.C:
			fmt.Println("slave_tick")
		}
	}
}

func handleInput(sl *Slave, nw *network.Network, message network.Message, slaveSend chan<- network.Message) {
	switch message.Response {
	case cl.Up:
		driver.SetMotorDirection(1)
		sl.MotorTimer.Reset(6 * time.Second)
	case cl.Down:
		driver.SetMotorDirection(-1)
		sl.MotorTimer.Reset(6 * time.Second)
	case cl.Stop:
		driver.SetMotorDirection(0)
		driver.SetDoorLamp(1)
		sl.MotorTimer.Stop()
		sl.DoorTimer.Reset(3 * time.Second)
	case cl.LightOnInner:
		driver.SetInnerPanelLamp(message.Content.(int), 1)
	case cl.LightOffInner:
		driver.SetInnerPanelLamp(message.Content.(int), 0)
	case cl.LightOnOuterUp:
		driver.SetOuterPanelLamp(1, message.Content.(int), 1)
	case cl.LightOffOuterUp:
		driver.SetOuterPanelLamp(1, message.Content.(int), 0)
	case cl.LightOnOuterDown:
		driver.SetOuterPanelLamp(-1, message.Content.(int), 1)
	case cl.LightOffOuterDown:
		driver.SetOuterPanelLamp(-1, message.Content.(int), 0)
	case cl.Connection:
		switch message.Content {
		case cl.Failed:
			//Assumes lost connection on timeout. This will be changed later
			send(nw.LocalIP, nw.LocalIP, cl.System, cl.SetMaster, slaveSend)
			sl.MasterID = nw.LocalIP
		}
	case cl.System:
		switch message.Content {
		case cl.JoinMaster:
			sl.StartupTimer.Stop()
			sl.MasterID = message.Sender
		}
	}
}

func send(masterID string, id string, response string, content interface{}, slaveSend chan<- network.Message) {
	if id == "" {
		id = network.CreateID(cl.Slave)
	}
	message := network.Message{
		Receiver: masterID,
		ID:       id,
		Response: response,
		Content:  content,
	}
	slaveSend <- message
}

func remoteInstall() {
	elevIP := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("This is a simple script for installing elevators on remote computers. \n Used on own responsibility, as this script is VERY simple and won't like input that contains any errors \n Good luck!")
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
