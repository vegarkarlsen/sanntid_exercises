package fsm

import (
	"Driver-go/elevio"
	"elevator/config"
	"elevator/orders"
	"log/slog"
	"time"
)

var elevator ElevatorState

func Fsm(
	buttonEventOutputChan 		chan<- elevio.ButtonEvent,
	clearOrdersChan				chan<- elevio.ButtonEvent,
	stateOutputChan 			chan<- ElevatorState,
	obstructionFaultChan		chan<- bool,
	elevatorBehaviorFaultChan	chan<- ElevatorBehavior,
	newOrdersChan 				<-chan [config.NumberFloors][3]bool) {

	slog.Info("\t[FSM SETUP]: Starting FSM, begin initializing of channels and elevator")

	buttonsChan 	:= make(chan elevio.ButtonEvent)
	floorSensorChan := make(chan int)
	obstructionChan := make(chan bool)
	doorTimerChan 	:= make(chan bool)

	go lightsHandler()
	go PollTimer(doorTimerChan)
	go checkClearedOrders(clearOrdersChan)
	go checkElevatorBehavior(elevatorBehaviorFaultChan)
	go elevio.PollButtons(buttonsChan)
	go elevio.PollFloorSensor(floorSensorChan)
	go elevio.PollObstructionSwitch(obstructionChan)
	slog.Info("\t[FSM SETUP]: Channels initialized")

	elevator = InitializeElevator()
	slog.Info("\t[FSM SETUP]: Elevator initialized")
	if elevator.Floor == -1 {
		onInitBetweenFloors()
	}

	slog.Info("\t[FSM SETUP]: Initialization complete, starting case handling")
	for {
		select {
		case obstruction := <-obstructionChan:
			slog.Info("\t[FSM Case]: Obstruction")
			onObstruction(obstruction)
			obstructionFaultChan <- elevator.Obstructed

		case buttonPress := <-buttonsChan:
			slog.Info("\t[FSM Case]: Button Press")
			onButtonPress(buttonPress, buttonEventOutputChan)

		case newFloor := <-floorSensorChan:
			slog.Info("\t[FSM Case]: New Floor", "floor", newFloor)
			onNewFloor(newFloor)

		case <-doorTimerChan:
			slog.Info("\t[FSM Case]: Door Timeout")
			onDoorTimeout()

		case ordersUpdate := <-newOrdersChan:
			slog.Info("\t[FSM Case]: New Orders")
			onOrdersUpdate(ordersUpdate)
		}
	}
}

func onInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Direction = elevio.MD_Down
	elevator.Behavior = EB_Moving
}

func onButtonPress(buttonPress elevio.ButtonEvent, sendToSyncChan chan<- elevio.ButtonEvent) {
	switch elevator.Behavior {
	case EB_DoorOpen:
		if ShouldClearImmediately(buttonPress.Floor, buttonPress.Button) {
			OpenDoor()
		} else {
			sendToSyncChan <- buttonPress
		}
	
	case EB_Idle:
		StartMotor()
		switch elevator.Behavior{
		case EB_DoorOpen:
			OpenDoor()
		default:
			sendToSyncChan <- buttonPress
		}

	default:
		sendToSyncChan <- buttonPress
	}
}

func onNewFloor(floor int) {
	elevator.Floor = floor
	elevio.SetFloorIndicator(elevator.Floor)
	switch elevator.Behavior {
	case EB_Moving:
		if ShouldStop() {
			ClearRequestAtCurrentFloor()
			StopMotor()
			OpenDoor()
		}
	}
}

func onDoorTimeout() {
	if elevator.Obstructed {
		StartTimer(config.DoorOpenTimeMs)
		return
	}

	timerActive = false
	switch elevator.Behavior {
	case EB_DoorOpen:
		directionBehavior := DecideMotorDirection()
		elevator.Behavior = directionBehavior.Behavior
		elevator.Direction = directionBehavior.Direction
		switch elevator.Behavior {
		case EB_DoorOpen:
			OpenDoor()
			ClearRequestAtCurrentFloor()
		default:
			CloseDoor()
			StartMotor()
		}
	default:
		CloseDoor()
		return
	}
}

func onObstruction(obstruction bool) {
	if obstruction {
		StopMotor()
		OpenDoor()
		elevator.Obstructed = true
	} else {
		elevator.Obstructed = false
		if elevator.Floor == -1 {
			onInitBetweenFloors()
		}
	}
}

func onOrdersUpdate(orders [config.NumberFloors][3]bool) {
	for i := 0; i < config.NumberFloors; i++ {
		for j := 0; j < 3; j++ {
			elevator.Orders[i][j] = orders[i][j]
		}
	}

	StartMotor()
	
	if elevator.Behavior == EB_DoorOpen{
		OpenDoor()
	}
}

func lightsHandler() {
	for {
		time.Sleep(time.Duration(config.LightUpdateTimeMs) * time.Millisecond)
        hallOrders := orders.GetHallOrders()
        cabOrders := orders.GetCabOrders(config.ElevatorId)
		for i := 0; i < config.NumberFloors; i++ {
            elevio.SetButtonLamp(elevio.BT_Cab, i, cabOrders[i])
			for j := 0; j < 2; j++ {
				elevio.SetButtonLamp(elevio.ButtonType(j), i, hallOrders[i][j])
			}
		}
	}
}

func checkClearedOrders(outputChan chan<- elevio.ButtonEvent) {
	previousOrders := elevator.Orders
	for {
		time.Sleep(time.Duration(config.CheckClearedOrdersTimeMs) * time.Millisecond)
		currentOrders := elevator.Orders

		for i := 0; i < config.NumberFloors; i++ {
			for j := 0; j < 3; j++ {
				if !currentOrders[i][j] && previousOrders[i][j]{
					outputChan <- elevio.ButtonEvent{Floor:i, Button:elevio.ButtonType(j)}
				}
			}
		}
		previousOrders = currentOrders
	}
}

func checkElevatorBehavior (outputChan chan<- ElevatorBehavior) {
	for {
		time.Sleep(time.Duration(config.ElevatorMalfunctionTimeMs) * time.Millisecond)
		outputChan <- elevator.Behavior
	}
}