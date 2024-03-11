package fsm

import (
	"Driver-go/elevio"
)

type ElevatorBehavior int

const (
	EB_Idle ElevatorBehavior = iota
	EB_Moving
	EB_DoorOpen
)

// Elevator state
type ElevatorState struct {
	Behavior   ElevatorBehavior
	Floor      int
	Direction  elevio.MotorDirection
	Requests   [][3]bool
	Obstructed bool
}

type DirectionBehaviorPair struct {
	Direction elevio.MotorDirection
	Behavior  ElevatorBehavior
}

func InitializeElevator(currentFloor int) ElevatorState {
	req := make([][3]bool, numFloors)
	for i := 0; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			req[i][j] = false
		}
	}

	return ElevatorState{EB_Idle, currentFloor, elevio.MD_Stop, req, false}
}

// Movement
var directionBehavior DirectionBehaviorPair

func StopMotor() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevator.Direction = elevio.MD_Stop
}

func StartMotor() {
	directionBehavior = DecideMotorDirection()
	elevio.SetMotorDirection(directionBehavior.Direction)
	elevator.Direction = directionBehavior.Direction
	elevator.Behavior = directionBehavior.Behavior
}

// Door
func OpenDoor() {
	elevator.Behavior = EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
	StartTimer(DoorOpenTime)
}
