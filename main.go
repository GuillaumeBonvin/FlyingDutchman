package main

import (
	"FlyingDutchman/internal"
	"fmt"
	"syscall"
)

func main() {
	fmt.Println("Welcome aboard, cabin boy !")
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
		case "d":
			fmt.Println(internal.FingerprintToPhrase("38:95:59:0a:7a:fc:8a:b4:4e:78:ae:8a:07:7f:5f:80:79:2d:39:04:f4:a3:27:e4:d2:90:63:bc:46:be:eb:4b"))

		default:
			fmt.Printf("Sorry, \"%s\" is not a functionnal command, please try again:\n", userResponse)
		}
	}
	fmt.Println("The glowing boat disappeared in the mist...")

}
