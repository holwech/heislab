package driver

/*
#include "elev.h"
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lpthread -lcomedi -lm
*/
import "C"

const cOuterPanelUp C.int = C.int(0)
const cOuterPanelDown C.int = C.int(1)
const cInnerPanel C.int = C.int(2)

type InnerOrder struct{
	Floor int
}
type OuterOrder struct{
	Floor, Direction int
}

func InitHardware(){
	C.elev_init()
}

func SetMotorDirection(direction int){
	C.elev_set_motor_direction(C.int(direction))
}

func SetOuterPanelLamp(direction, floor, value int){
	if direction == 1 {
		C.elev_set_button_lamp(cOuterPanelUp,C.int(floor),C.int(value))
	}else if direction == -1{
		C.elev_set_button_lamp(cOuterPanelDown,C.int(floor),C.int(value))
	}
}
func SetInnerPanelLamp(floor,value int) {
	C.elev_set_button_lamp(cInnerPanel,C.int(floor),C.int(value))
}

func SetFloorIndicatorLamp(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func SetDoorLamp(value int){
	C.elev_set_door_open_lamp(C.int(value))
}

func SetStopLamp(value int){
	C.elev_set_stop_lamp(C.int(value))
}

func ListenInnerPanel() <-chan InnerOrder{
	orderChan := make(chan InnerOrder)
	go func(){
		buttonPressed := [4]bool {false,false,false,false}
		for{
			for floor := 0; floor < 4; floor++{
				if C.elev_get_button_signal(cInnerPanel,C.int(floor)) != 0{
					if buttonPressed[floor] == false{
						var order InnerOrder
						order.Floor = floor
						orderChan <- order	
						buttonPressed[floor] = true
					}
				}else{
					buttonPressed[floor] = false
				}
			} 
		}
	}()
	return orderChan
}

func ListenOuterPanel() <-chan OuterOrder{
	orderChan := make(chan OuterOrder)
	go func(){
		buttonPressedUp := [4]bool {false,false,false,false}
		buttonPressedDown := [4]bool {false,false,false,false}
		for{
			for floor := 0; floor < 4; floor++{
				if C.elev_get_button_signal(cOuterPanelUp,C.int(floor)) != 0{
					if buttonPressedUp[floor] == false{				
						var order OuterOrder
						order.Floor = floor
						order.Direction = 1
						orderChan <- order
						buttonPressedUp[floor] = true
					}
				}else{
					buttonPressedUp[floor] = false
				}
				if C.elev_get_button_signal(cOuterPanelDown,C.int(floor)) != 0{
					if buttonPressedDown[floor] == false{
						var order OuterOrder
						order.Floor = floor
						order.Direction = -1
						orderChan <- order	
						buttonPressedDown[floor] = true
					}
				} else{
					buttonPressedDown[floor] = false
				}
			} 
		}
	}()
	return orderChan
}

func ListenFloorSensor() <-chan int{
	floorChan := make(chan int)
	go func(){
		prevFloor := -2;
		for{
			currentFloor := int(C.elev_get_floor_sensor_signal())
			if currentFloor != prevFloor{
				floorChan <- currentFloor
				prevFloor = currentFloor
			}
		}
	}()
	return floorChan
}


