package master

import network

type Behaviour int
const (
	Idle Behaviour = iota
	Moving
	Stopped
)

type ElevatorState struct{
	Floor int
	Direction int
	CurrentBehaviour Behaviour
	InnerOrders [4] bool 
}

//Listen to inputs from slaves and send actions back
func master(recv chan communication.CommData,send chan communication.CommData){
	//messageChan, statusChan := InitNetwork()
	//sendChan := network get send chan

	
	elevatorStates := make(map[string]ElevatorState)
	outerOrdersUp := []bool{false,false,false,false}
	outerOrdersDown := []bool{false,false,false,false}
	

	testState := ElevatorState{1,0, Idle}
	elevators["localhost"] =testState

	for{
		select{
		case message := <- messageChan:
			//Decode message, do corresponding action
			switch commData.DataType{
			case "INNER":
				order := commData.DataValue.(driver.InnerOrder)
				if elevatorStates["localhost"].Floor != order.floor{
					var elevator = elevators["localhost"]
					elevator.InnerOrders[order.floor-1] = true
					elevators["localhost"] = elevator
					command, hasCommand := orders.GetCommand(elevatorStates,outerOrdersUp,outerOrdersDown)
					if hasCommand{
						send <- communication.CommData{
							DataType = command,
						}
					}
				}
			case "OUTER":
				order := commData.DataValue.(driver.OuterOrder)
				if(order.Direction == 1){
					outerOrdersUp[order.Floor-1] = true
				}else{
					outerOrdersDown[order.Floor-1] = true
				}
				command, hasCommand := orders.GetCommand(elevatorStates,outerOrdersUp,outerOrdersDown)
				if hasCommand{
					send <- communication.CommData{
						DataType = command,
					}
				}
			case "FLOOR":
				floor := commData.DataValue.(int)
				var elevator = elevators["localhost"]
				elevator.Floor = floor
				elevators["localhost"] = elevator

			}
		case connStatus := <- statusChan:
			update connected
		}

	}

}
