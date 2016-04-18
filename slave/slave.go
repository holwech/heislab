package slave

import (
	"time"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
)

type Slave struct {
	DoorTimer, StartupTimer, MotorTimer *time.Timer
	MasterID, EngineState string
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
	sl.StartupTimer.Reset(50 * time.Millisecond)
	send(nw.LocalIP, "", cl.Startup, time.Now(), slaveSend)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case innerOrder := <-innerChan:
			send(sl.MasterID, "", cl.InnerOrder, innerOrder, slaveSend)
		case outerOrder := <-outerChan:
			send(sl.MasterID, "", cl.OuterOrder, outerOrder, slaveSend)
		case newFloor := <-floorChan:
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
		case <- sl.MotorTimer.C:
			sl.EngineState = cl.EngineFail
			send(sl.MasterID, "", cl.System, cl.EngineFail, slaveSend)
		case <- ticker.C:
			fmt.Println("Slave alive")
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
			send(nw.LocalIP, "", cl.SetMaster, time.Now(), slaveSend)
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
