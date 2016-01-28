package main

import "github.com/holwech/heislab/driver"

func main(){
	driver.InitHardware()
	driver.SetButtonLamp(1,1,1)
}