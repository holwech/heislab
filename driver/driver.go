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
		C.elev_set_button_lamp(cOuterPanelUp,castCfloor(floor),C.int(value))
	}else if direction == -1{
		C.elev_set_button_lamp(cOuterPanelDown,castCfloor(floor),C.int(value))
	}
}
func SetInnerPanelLamp(floor,value int) {
	C.elev_set_button_lamp(cInnerPanel,castCfloor(floor),C.int(value))
}

func SetFloorIndicatorLamp(floor int){
	C.elev_set_floor_indicator(castCfloor(floor))
}

func SetDoorLamp(value int){
	C.elev_set_door_open_lamp(C.int(value))
}

func SetStopLamp(value int){
	C.elev_set_stop_lamp(C.int(value))
}

func ReadInnerPanel() <-chan InnerOrder{
	orderChan := make(chan InnerOrder)
	go func(){
		buttonPressed := [4]bool {false,false,false,false}
		for{
			for floor := 1; floor <= 4; floor++{
				if C.elev_get_button_signal(cInnerPanel,castCfloor(floor)) != 0{
					if buttonPressed[floor-1] == false{
						var order InnerOrder
						order.Floor = floor
						orderChan <- order	
						buttonPressed[floor-1] = true
					}
				}else{
					buttonPressed[floor-1] = false
				}
			} 
		}
	}()
	return orderChan
}

func ReadOuterPanel() <-chan OuterOrder{
	orderChan := make(chan OuterOrder)
	go func(){
		buttonPressedUp := [4]bool {false,false,false,false}
		buttonPressedDown := [4]bool {false,false,false,false}
		for{
			for floor := 1; floor <= 4; floor++{
				if C.elev_get_button_signal(cOuterPanelUp,castCfloor(floor)) != 0{
					if buttonPressedUp[floor-1] == false{				
						var order OuterOrder
						order.Floor = floor
						order.Direction = 1
						orderChan <- order
						buttonPressedUp[floor-1] = true
					}
				}else{
					buttonPressedUp[floor-1] = false
				}
				if C.elev_get_button_signal(cOuterPanelDown,castCfloor(floor)) != 0{
					if buttonPressedDown[floor-1] == false{
						var order OuterOrder
						order.Floor = floor
						order.Direction = -1
						orderChan <- order	
						buttonPressedDown[floor-1] = true
					}
				} else{
					buttonPressedDown[floor-1] = false
				}
			} 
		}
	}()
	return orderChan
}

func ReadFloorSensor() <-chan int{
	floorChan := make(chan int)
	go func(){
		for{
			floorChan <- castFloor(C.elev_get_floor_sensor_signal())
		}
	}()
	return floorChan
}

func castCfloor(floor int) C.int{
	return C.int(floor-1)
}

func castFloor(cFloor C.int) int{
	sensorVal := int(cFloor)
	floor := 0

	if sensorVal == -1{
		floor = -1
	} else{
		floor = sensorVal +1
	} 
	return floor
}
