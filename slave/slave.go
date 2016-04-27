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
	sl.MasterID = cl.All
}

func initSlave() *Slave {
	sl := new(Slave)
	sl.Init()
	return sl
}

func Run() {
	innerChan, outerChan, floorChan := driver.InitElevator()
	nw, _ := network.InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	master.InitMaster()
	sl := initSlave()
	receive, send := nw.Channels()
	time.Sleep(50 * time.Millisecond)
	ticker := time.NewTicker(8 * time.Second)

	for {
		select {
		case innerOrder := <-innerChan:
			network.Send(sl.MasterID, cl.Slave, cl.InnerOrder, innerOrder, send)
		case outerOrder := <-outerChan:
			network.Send(sl.MasterID, cl.Slave, cl.OuterOrder, outerOrder, send)
		case newFloor := <-floorChan:
			fmt.Printf("Floor: %d\n", newFloor)
			network.Send(sl.MasterID, cl.Slave, cl.Floor, newFloor, send)
			if newFloor != -1 {
				sl.MotorTimer.Reset(6 * time.Second)
				if sl.EngineState == cl.EngineFail {
					sl.EngineState = cl.EngineOK
					network.Send(sl.MasterID, cl.Slave, cl.System, cl.EngineOK, send)
				}
			}
		case <-sl.DoorTimer.C:
			driver.SetDoorLamp(0)
			network.Send(sl.MasterID, cl.Slave, cl.DoorClosed, "", send)
		case message := <-receive:
			handleInput(sl, nw, message, send)
		case <-sl.MotorTimer.C:
			sl.EngineState = cl.EngineFail
			network.Send(sl.MasterID, cl.Slave, cl.System, cl.EngineFail, send)
		case <-ticker.C:
			fmt.Println("slave_tick")

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
			network.PrintMessage(&message)
		}
	}
}
