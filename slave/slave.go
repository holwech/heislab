package slave

import (
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/driver"
	"github.com/holwech/heislab/master"
	"github.com/holwech/heislab/network"
	"time"
)

func Run(backup bool) {
	innerChan, outerChan, floorChan := driver.InitElevator()
	nw := network.InitNetwork(cl.SReadPort, cl.SWritePort, cl.Slave)
	DoorTimer := time.NewTimer(time.Second)
	DoorTimer.Stop()
	MotorTimer := time.NewTimer(time.Second)
	MotorTimer.Stop()
	go master.Run(backup)
	receive, send := nw.Channels()
	time.Sleep(50 * time.Millisecond)

	for {
		select {
		case innerOrder := <-innerChan:
			network.Send(cl.All, cl.Slave, cl.InnerOrder, innerOrder, send)
		case outerOrder := <-outerChan:
			network.Send(cl.All, cl.Slave, cl.OuterOrder, outerOrder, send)
		case newFloor := <-floorChan:
			network.Send(cl.All, cl.Slave, cl.Floor, newFloor, send)
			if newFloor != -1 {
				MotorTimer.Reset(6 * time.Second)
			}
		case <-DoorTimer.C:
			driver.SetDoorLamp(0)
			network.Send(cl.All, cl.Slave, cl.DoorClosed, cl.Slave, send)
		case <-MotorTimer.C:
			network.Send(cl.All, cl.Slave, cl.System, cl.EngineFail, send)
		case message := <-receive:
			switch message.Response {
			case cl.Up:
				driver.SetMotorDirection(1)
				MotorTimer.Reset(6 * time.Second)
			case cl.Down:
				driver.SetMotorDirection(-1)
				MotorTimer.Reset(6 * time.Second)
			case cl.Stop:
				driver.SetMotorDirection(0)
				driver.SetDoorLamp(1)
				MotorTimer.Stop()
				DoorTimer.Reset(3 * time.Second)
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
					network.PrintMessage(&message, "Slave received message")
				}
			}
		}
	}
}
