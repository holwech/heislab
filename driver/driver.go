package driver

/*
#include "elev.h"
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lpthread -lcomedi -lm
*/
import "C"

const numFloors = 4

type Order struct{
	Type string
	Floor, Direction int
}

func InitHardware(){
	C.elev_init()
}

func SetMotorDirection(direction int){
	C.elev_set_motor_direction(C.int(direction))
}

func SetOuterPanelLamp(direction, floor, value int){
	if direction == 0 {
		C.elev_set_button_lamp(C.int(0),C.int(floor),C.int(value))
	}else{
		C.elev_set_button_lamp(C.int(1),C.int(floor),C.int(value))
	}
}
func SetInnerPanelLamp(floor,value int) {
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

func ReadOrders(orderChan chan Order){
	for{
		for floor := 0; floor < numFloors; floor++{
			if C.elev_get_button_signal(C.int(2), C.int(floor)) != 0{
				var order Order
				order.Type = "inner"
				order.Floor = floor+1
				orderChan <- order
			}
		}
		for floor := 0; floor < numFloors; floor++{
			for direction := 0; direction < 2; direction++{
				if C.elev_get_button_signal(C.int(direction), C.int(floor)) != 0{
					var order Order
					order.Type = "outer"
					order.Floor = floor+1
					//Make directions from 1/0 (from elev_get_button_signal) to -1/1
					if direction == 0{
						order.Direction = 1
					}else  {
						order.Direction = -1
					}
					orderChan <- order
				}
			}
		}
	}	
}

func ReadFloorSensor(floorChan chan int){
	for{
		floorChan <- int(C.elev_get_floor_sensor_signal())
	}
}
