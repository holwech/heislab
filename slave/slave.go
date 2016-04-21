package slave

import (
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
	"time"
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
	nw := network.InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	master.InitMaster()
	sl := initSlave()
	receive, send := nw.Channels()
	time.Sleep(50 * time.Millisecond)
	sendMsg(nw.LocalIP, "", cl.System, cl.Startup, send)
	ticker := time.NewTicker(3 * time.Second)

	sl.StartupTimer.Reset(50 * time.Millisecond)
	for {
		select {
		case innerOrder := <-innerChan:
			sendMsg(sl.MasterID, "", cl.InnerOrder, innerOrder, send)
		case outerOrder := <-outerChan:
			sendMsg(sl.MasterID, "", cl.OuterOrder, outerOrder, send)
		case newFloor := <-floorChan:
			fmt.Printf("Floor: %d\n", newFloor)
			sendMsg(sl.MasterID, "", cl.Floor, newFloor, send)
			if newFloor != -1 {
				sl.MotorTimer.Reset(6 * time.Second)
				if sl.EngineState == cl.EngineFail {
					sl.EngineState = cl.EngineOK
					sendMsg(sl.MasterID, "", cl.System, cl.EngineOK, send)
				}
			}
		case <-sl.DoorTimer.C:
			driver.SetDoorLamp(0)
			sendMsg(sl.MasterID, "", cl.DoorClosed, "", send)
		case message := <-receive:
			handleInput(sl, nw, message, send)
		case <-sl.StartupTimer.C:
			sendMsg(nw.LocalIP, "", cl.System, cl.SetMaster, send)
			sl.MasterID = nw.LocalIP
		case <-sl.MotorTimer.C:
			sl.EngineState = cl.EngineFail
			sendMsg(sl.MasterID, "", cl.System, cl.EngineFail, send)
		case <-ticker.C:
			//fmt.Println("slave_tick")

		}
	}
}

func handleInput(sl *Slave, nw *network.Network, message network.Message, send chan<- network.Message) {
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
			if sl.MasterID != nw.LocalIP {
				sendMsg(nw.LocalIP, "", cl.System, cl.SetMaster, send)
				sl.MasterID = nw.LocalIP
			}
		}
	case cl.System:
		switch message.Content {
		case cl.JoinMaster:
			sl.StartupTimer.Stop()
			sl.MasterID = message.Sender
		}
	}
}

func sendMsg(masterID string, id string, response string, content interface{}, send chan<- network.Message) {
	if id == "" {
		id = network.CreateID(cl.Slave)
	}
	message := network.Message{
		Receiver: masterID,
		ID:       id,
		Response: response,
		Content:  content,
	}
	send <- message
}
