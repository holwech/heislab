package driver

/*
#include "elev.h"
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lpthread -lcomedi -lm
*/
import "C"

const numFloors = 4


func InitHardware(){
	C.elev_init()
}

func SetMotorDirection(dirn int){
	C.elev_set_motor_direction(C.int(dirn))
}

func SetOuterPanelLamp(direction, floor, value int){
	if direction == 0 {
		C.elev_set_button_lamp(C.int(0),C.int(floor),C.int(value))
	}else{
		C.elev_set_button_lamp(C.int(1),C.int(floor),C.int(value))
	}
}
func SetInnerPanelLamp(btn,floor,value int) {
	C.elev_set_button_lamp(C.int(2),C.int(floor),C.int(value))
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

func ReadInnerPanel(btnChan chan int){
	for{
		for floor := 0; floor < numFloors; floor++{
			if C.elev_get_button_signal(C.int(2), C.int(floor)) != 0{
				btnChan <- floor
			}
		}
	}
}


func ReadOuterPanel(btnChan chan int, directionChan chan int){
	for{
		for floor := 0; floor < numFloors; floor++{
			for direction := 0; direction < 2; direction++{
				if C.elev_get_button_signal(C.int(direction), C.int(floor)) != 0{
					btnChan <- floor
					directionChan <- direction
				}
			}
		}
	}
}

func ReadFloorSignal(floorChan chan int){
	for{
		floorChan <- int(C.elev_get_floor_sensor_signal())
	}
}
