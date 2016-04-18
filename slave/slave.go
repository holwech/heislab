package main

import (
	"time"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
)

type Slave struct {
	DoorTimer, StartupTimer, MotorTimer *time.Timer
	MasterID string
}

func (sl *Slave) Init() {
	sl.DoorTimer = time.NewTimer(time.Second)
	sl.DoorTimer.Stop()
	sl.StartupTimer = time.NewTimer(time.Second)
	sl.StartupTimer.Stop()
	sl.MotorTimer = time.NewTimer(time.Second)
	sl.MotorTimer.Stop()
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
	sl.StartupTimer.Reset(50 * time.Millisecond)
	send(nw.LocalIP, cl.Startup, time.Now(), slaveSend)
	for {
		select {
		case innerOrder := <-innerChan:
			send(sl.MasterID, cl.InnerOrder, innerOrder, slaveSend)
		case outerOrder := <-outerChan:
			send(sl.MasterID, cl.OuterOrder, outerOrder, slaveSend)
		case newFloor := <-floorChan:
			send(sl.MasterID, cl.Floor, newFloor, slaveSend)
		case <- sl.DoorTimer.C:
			driver.SetDoorLamp(0)
			send(sl.MasterID, cl.DoorClosed, "", slaveSend)
		case message := <-slaveReceive:
			handleInput(sl, nw, message, slaveSend)
		case <- sl.StartupTimer.C:
			send(nw.LocalIP, cl.SetMaster, time.Now(), slaveSend)
			sl.MasterID = nw.LocalIP
		}
	}
}

func handleInput(sl *Slave, nw *network.Network, message network.Message, slaveSend chan<- network.Message) {
	switch message.Response {
	case cl.Up:
		driver.SetMotorDirection(1)
	case cl.Down:
		driver.SetMotorDirection(-1)
	case cl.Stop:
		driver.SetMotorDirection(0)
		driver.SetDoorLamp(1)
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
		switch message.Content{
		case cl.Failed:
			//Assumes lost connection on timeout. This will be changed later
			send(nw.LocalIP, cl.SetMaster, time.Now(), slaveSend)
			sl.MasterID = nw.LocalIP
		}
	case cl.System:
		switch message.Content{
		case cl.JoinMaster:
			sl.StartupTimer.Stop()
			sl.MasterID = message.Sender
		}
	}
}

func send(masterID string, response string, content interface{}, slaveSend chan<- network.Message) {
	message := network.Message{
		Receiver: masterID,
		ID:       network.CreateID(cl.Slave),
		Response: response,
		Content:  content,
	}
	slaveSend <- message
}

func main() {
	Run()
}
