package scheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/holwech/heislab/cl"
	"github.com/holwech/heislab/network"
	"io/ioutil"
)

type ElevatorState struct {
	Floor            int
	Direction        int
	InnerOrders      [cl.Floors]bool
	DoorClosed			 bool
	EngineFail			 bool
	PingTime				 time.Time
}

type System struct {
	Elevators           map[string]ElevatorState
	OrdersUp					  [cl.Floors]bool
	OrdersDown				  [cl.Floors]bool
	IsActive						bool
}

func NewSystem(localIP string) *System{
	sys := new(System)
	sys.Elevators = make(map[string]ElevatorState)
	elevator := ElevatorState{Floor: 0, Direction: 0}
	sys.Elevators[localIP] = elevator
	sys.IsActive = true
	return sys
}

func (sys *System) CheckTimeout(localIP string) bool {
	disconnect := false
	for elevIP, elevState := range sys.Elevators {
		diff := time.Sub(elevState.PingTime)
		if 300 * time.Millisecond < diff && elevIP != localIP {
			delete(sys.Elevators, elevIP)
			disconnect = true
			fmt.Println(elevIP, " disconnected")
		}
	}
	return disconnect
}

func (sys *System) UpdatePingTime(elevIP string) bool {
	newElev := false
	if _, exists := sys.Elevators[elevIP], !exists {
		sys.Elevators[elevIP] = ElevatorState{}
		newElev = true
		fmt.Println(elevIP, " connected")
	}
	elevator := &sys.Elevators[elevIP]
	elevator.PingTime = time.Now()
	return newElev
}

func (sys *System) MergeSystems(message *network.Message) {
	sys2 := backupToSystem(&message)
	for i, val := range sys.OrdersUp {
		sys.OrdersUp[i] = (sys.OrdersUp[i] || sys2.OrdersUp[i])
	}
	for i, val := range sys.OrdersDown {
		sys.OrdersDown[i] = (sys.OrdersDown[i] || sys2.OrdersDown[i])
	}
	for elevIP, elevState := range sys2.Elevators {
		sys.Elevators[elevIP] = elevState
	}
}

func (sys *System) SetActive(localIP string) {
	for elevIP, _ := range sys.Elevators {
		if elevIP < localIP {
			sys.IsActive = false
		}
	}
	sys.IsActive = true
}

func (sys *System) Recover(localIP string) {
	backup = ReadFromFile()
	sys = scheduler.NewSystem(nwMaster.LocalIP)
	elevator := &sys.Elevators[localIP]
	elevatorBackup := backup.Elevators[localIP]
	elevator.InnerOrders = elevatorBackup.InnerOrders
	if len(backup.Elevators) == 1 {
		 sys.OrdersUp = backup.OrdersUp
		 sys.OrdersDown = backup.OrdersDown
	}
}

func (sys *System) CreateBackup() map[string]interface{} {
	var buffer bytes.Buffer
	var backup map[string][]byte
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(*sys)
	printError("Encode error in CreateBackup:", err)
	backup[cl.Backup] = buffer.Bytes()
	return backup
}

func backupToSystem(message *network.Message) *System {
	var sys *System
	printError("ReadFile error in backupToSystem:", err)
	buffer := bytes.NewBuffer(message.Content[cl.Backup])
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(sys)
	printError("Decode error in backupToSystem:", err)
	return sys
}

func (sys *System) RefreshAllElevators() []network.Message {
	var orders []network.Message
	for elevIP, _ := range sys.Elevators {
		innerOrders := sys.RefreshInnerLights(elevIP)
		orders = append(orders, innerOrders...)
	}
	outerOrders := sys.RefreshOuterLights()
	orders = append(orders, outerOrders...)
	return orders
}

func (sys *System) RefreshInnerLights(elevIP string) []network.Message {
	var orders []network.Message
	elevator := sys.Elevators[elevIP]
	for _, lightOn := range elevator.InnerOrders {
		response := cl.LightOffInner
		if lightOn {
			response = cl.LightOnInner
		}
		message = network.Message{ "", elevIP, "", response, map[string]interface{}{cl.Floor: i}}
		orders = append(orders, message)
	}
	return orders
}

func (sys *System) RefreshOuterLights() []network.Message {
	var orders []network.Message
	for _, lightOn := range sys.OrdersUp {
		response := cl.LightOffOuterUp
		if lightOn {
			response = cl.LightOnOuterUp
		}
		message := network.Message{ "", cl.All, "", response, map[string]interface{}{cl.Floor: i}}
		orders = append(orders, message)
		response = cl.LightOffOuterDown
		if lightOn {
			response = cl.LightOnOuterDown
		}
		message = network.Message{ "", cl.All, "", response, map[string]interface{}{cl.Floor: i}}
		orders = append(orders, message)
	}
	return orders
}

func (sys *System) Print() {
	fmt.Println("------------------------------")
	fmt.Println("SYSTEM ACTIVE:", sys.IsActive)
	fmt.Println("Orders up:", sys.OrdersUp, "Orders down:", sys.OrdersDown)
	for elevIP, elevState := range sys.Elevators {
		fmt.Println(elevIP, "---------------")
		fmt.Println("Floor:", elevState.Floor,
								"Direction:", elevState.Direction,
								"DoorClosed:", elevState.DoorClosed,
								"EngineFail:", elevState.EngineFail)
		fmt.Println("Inner orders:", elevState.InnerOrders)
		fmt.Println("------------------------------")
	}
}

func (sys *System) WriteToFile() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(*sys)
	printError("Encode error: ", err)
	err = ioutil.WriteFile("tmp", buffer.Bytes(), 0644)
	printError("WriteFile error: ", err)
	fmt.Println("BACKUP: Backup stored")
}

func ReadFromFile() *System {
	var sys *System
	file, err := ioutil.ReadFile("tmp")
	printError("ReadFile error: ", err)
	buffer := bytes.NewBuffer(file)
	fmt.Println("BACKUP: Reading from file...")
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(sys)
	printError("Decode error: ", err)
	return sys
}

func printError(comment string, err error) {
	if err != nil {
		fmt.Println(comment, err)
		fmt.Println(err.Error())
	}
}
