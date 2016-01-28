package driver

/*
#include "elev.h"
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lpthread -lcomedi -lm
*/
import "C"
//import "fmt"
func InitHardware(){
	C.elev_init()
}

func SetButtonLamp(button, floor, value int){
	C.elev_set_button_lamp(C.int(button),C.int(floor),C.int(value))
}

/*
func SetMotorDirection(dirn int);
	C.elev_set_motor_direction(C.int(dirn));
}
*/
/*
func elev_set_floor_indicator(int floor);
func elev_set_door_open_lamp(int value);
func elev_set_stop_lamp(int value);


func GetButtonSignal(button int, floor int){
	return int(C.elev_get_button_signal(C.int(button), C.int(floor)))
}
int elev_get_floor_sensor_signal(void);
int elev_get_stop_signal(void);
int elev_get_obstruction_signal(void);
*/