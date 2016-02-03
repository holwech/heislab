package types

type InnerOrder struct{
	//elevatorID string
	Floor int
}

type OuterOrder struct{
	//elevatorID string
	Floor, Direction int
}

type ElevatorState struct{
	Floor, Direction, Request int
	IsInFloor bool
}
