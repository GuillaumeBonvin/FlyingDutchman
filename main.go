package main

import (
	"fmt"
)

func main() {
	fmt.Println("Something appears out of the deep sea... The Flying Dutchman!")
	var userResponse string
	var userStateDefined = false
	var userIsDone = false

	for !userStateDefined && !userIsDone {
		fmt.Println("Are you sender or receiver of the file? ('s'/'r')\nYou can stop the program by typing quit ('q')")
		fmt.Scanln(&userResponse)

		switch userResponse {
		case "s", "send":
			userStateDefined = true
			fmt.Println("Preparing to send...\n Please enter file path:")
			var filepath string
			fmt.Scanln(&filepath)
			Sender(filepath)
		case "r", "receive":
			userStateDefined = true
			fmt.Println("Preparing to receive...\n Please indicate output path:")
			var outpath string
			fmt.Scanln(&outpath)
			Receiver(outpath)
		case "q", "quit":
			userStateDefined = true
			userIsDone = true
			break

		default:
			fmt.Printf("Sorry, \"%s\" is not a functionnal command, please try again:\n", userResponse)
		}
	}
	fmt.Println("The glowing boat disappeared in the mist...")

}
