package main

import (
	"FlyingDutchman/internal"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v3"
	"github.com/sacOO7/gowebsocket"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func Receiver(outputPath string) {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.
	fmt.Println("enter remote sdp:")
	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:       []string{"turn:turn.flying-dut.ch:3478", "stun:stun.flying-dut.ch:3478"},
				Username:   "captain",
				Credential: "Axp2oSr56d5",
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open.\n", d.Label(), d.ID())

		})

		// Register text message handling
		rebuiltFile := []byte{}

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			//fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			//fmt.Println("file received!")

			rebuiltFile = append(rebuiltFile, msg.Data[:]...)

			// Convert byte array to file
			err := ioutil.WriteFile(outputPath, rebuiltFile, 0644)
			if err != nil {
				panic(err)
			}
		})
	})

	/////// START OF WS SIGNALING ///////////

	type Message struct {
		Type    string
		Success bool
		Offer   string
		Answer  string
		Name    string
		Sender  string
	}

	var name string
	fmt.Println("Enter your name:")
	fmt.Scanln(&name)

	var remote string

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New("ws://signal.flying-dut.ch:9090")

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")

		ans := Message{Type: "login", Name: name}
		b, err := json.Marshal(ans)
		if err != nil {
			panic(err)
		}
		socket.SendBinary(b)
	}

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Recieved connect error ", err)
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		//log.Println("Recieved message " + message)
		var m Message

		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			panic(err)
		}
		switch m.Type {
		case "login":
			if m.Success == true {
				log.Println("Login success")

			} else {
				log.Println("Login failed")
			}
		case "offer":
			remote = m.Name

			log.Println("Received offer from " + m.Name)
			fmt.Println("Do you want to accept the offer? ('yes'/'no')")
			var accept string
			fmt.Scanln(&accept)

			var stateDefined = false
			for !stateDefined {

				switch accept {
				case "yes", "y":

					stateDefined = true

					// Set the offer as remote description
					var encodedOffer = m.Offer
					offer := webrtc.SessionDescription{}
					internal.Decode(encodedOffer, &offer)

					err = peerConnection.SetRemoteDescription(offer)
					if err != nil {
						panic(err)
					}

					// Create an answer
					answer, err := peerConnection.CreateAnswer(nil)
					if err != nil {
						panic(err)
					}

					// Create channel that is blocked until ICE Gathering is complete
					gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

					// Sets the LocalDescription, and starts our UDP listeners
					err = peerConnection.SetLocalDescription(answer)
					if err != nil {
						panic(err)
					}

					// Block until ICE Gathering is complete, disabling trickle ICE
					// we do this because we only can exchange one signaling message
					// in a production application you should exchange ICE Candidates via OnICECandidate
					<-gatherComplete

					encodedAnswer := internal.Encode(*peerConnection.LocalDescription())

					ans := Message{Type: "answer", Name: remote, Answer: encodedAnswer}
					b, err := json.Marshal(ans)
					if err != nil {
						panic(err)
					}
					socket.SendBinary(b)
					log.Println("Sending answer to " + remote)

					// notify and close connection
					msg := Message{Type: "leave", Name: remote}
					c, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					socket.SendBinary(c)

					socket.Close()

				case "no", "n":
					stateDefined = true

					msg := Message{Type: "reject", Name: remote}
					c, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					socket.SendBinary(c)
					fmt.Println("Waiting for another offer...")
				default:
					fmt.Println("Command " + accept + " is not defined, please type 'yes' to accept offer, 'no' to decline:")
					fmt.Scanln(&accept)
				}
			}
		}
	}

	socket.OnBinaryMessage = func(data []byte, socket gowebsocket.Socket) {
		log.Println("Recieved binary data ", data)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Recieved ping " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Recieved pong " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}

	socket.Connect()

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
	/////// END OF WS SIGNALING ///////////

}
