package driver

import (
	"testing"
	"os"
	"fmt"
	"time"
)




func TestListenFloor(t *testing.T){
	currentFloor := ListenFloorSensor()	
	fmt.Printf("Current floor %d \n", <-currentFloor)
}

func TestSetMotor(t *testing.T){
	currentFloor := ListenFloorSensor()
	SetMotorDirection(-1);
	for{
		floor := <- currentFloor
		if floor > 0{
			SetMotorDirection(0)
			break;
		}
	}
}

func TestOuterLights(t *testing.T){
	SetOuterPanelLamp(1,1,1)
	SetOuterPanelLamp(-1,4,1)
	time.Sleep(3*time.Second)
	SetOuterPanelLamp(1,1,0)
	SetOuterPanelLamp(-1,4,0)
}

func TestInnerLights(t *testing.T){
	SetInnerPanelLamp(1,1)
	SetInnerPanelLamp(4,1)
	time.Sleep(3*time.Second)
	SetInnerPanelLamp(1,0)
	SetInnerPanelLamp(4,0)
}

func TestFloorIndicator(t *testing.T){
	for i := 1; i <= 4; i++{
		SetFloorIndicatorLamp(i)
		time.Sleep(500*time.Millisecond)
	}
}

func TestDoorLamp(t *testing.T){
	SetDoorLamp(1)
	time.Sleep(1*time.Second)
	SetDoorLamp(0)
}

func TestStopLamp(t *testing.T){
	SetStopLamp(1)
	time.Sleep(1*time.Second)
	SetStopLamp(0)
}

func TestListenInner(t *testing.T){
	innerOrders := ListenInnerPanel()
	fmt.Println(<-innerOrders)
	fmt.Println(<-innerOrders)
}

func TestListenOuter(t *testing.T){
	innerOrders := ListenOuterPanel()
	fmt.Println(<-innerOrders)
	fmt.Println(<-innerOrders)
}


func TestMain(m *testing.M) {
	InitHardware()

	os.Exit(m.Run())
}