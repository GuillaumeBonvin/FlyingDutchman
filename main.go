package main

import (
	"fmt"
	"log"
	"syscall"
)

var keepExe = true

func recoverMain() {
	if r := recover(); r != nil {
		log.Println("An error occured!\nRecovered from ", r)
		keepExe = true
	}
}
func choseState() {
	defer recoverMain()

	fmt.Println("\nWelcome aboard, cabin boy !")
	var userResponse string
	var userStateDefined = false
	var userIsDone = false

	for !userStateDefined && !userIsDone {
		fmt.Println("Are you sender or receiver of the file? ('s'/'r')\nYou can stop the program by typing quit ('q')")
		fmt.Scanln(&userResponse)

		switch userResponse {
		case "s", "send":
			userStateDefined = true
			fmt.Println("Preparing to send...")
			Sender()
			userIsDone = true

		case "r", "receive":
			userStateDefined = true
			fmt.Println("Preparing to receive...")
			Receiver()
			userIsDone = true

		case "q", "quit":
			userStateDefined = true
			userIsDone = true
			syscall.Exit(0)

		default:
			fmt.Printf("Sorry, \"%s\" is not a functionnal command, please try again:\n", userResponse)
		}
	}
	fmt.Println("The glowing boat disappeared in the mist...")

}
func main() {
	for keepExe {
		keepExe = true
		choseState()
	}
}
