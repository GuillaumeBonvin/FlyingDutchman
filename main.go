package main

import (
	"fmt"
)

func main() {
	fmt.Println("Welcome to the Flying Dutchman!\nAre you sender or receiver of the file? ('s'/'r')")
	var userResponse string
	var userStateDefined = false

	for !userStateDefined {
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
		case "t", "test":

		default:
			fmt.Printf("Sorry, \"%s\" is not a functionnal command, please try again:\n", userResponse)
		}
	}

}
