package main

import (
	"FlyingDutchman/internal"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v3"
	"github.com/sacOO7/gowebsocket"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func Receiver() {

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

	// Generate your personal certificate passphrase
	tlsFingerprints, err := peerConnection.GetConfiguration().Certificates[0].GetFingerprints()
	fingerprint := internal.FingerprintToString(tlsFingerprints[0])
	localPassphrase := internal.FingerprintToPhrase(fingerprint)
	fmt.Println("Your passphrase is: " + localPassphrase)

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState.String() == "disconnected" {
			fmt.Println("Remote user disconnected: Taking you back to main menu.")
			peerConnection.Close()
			main()
		}
	})

	////////////////////////////////////////// FILE EXCHANGE PROTOCOL //////////////////////////////////////////////////

	type Exchange struct {
		Type     string
		FileName string
		FileSize int64
		Hash     []byte
		Data     []byte
	}

	var outputPath string
	var fileHash []byte

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {

		log.Printf("New DataChannel %s %dataChannel\n", dataChannel.Label(), dataChannel.ID())

		// Register channel opening handling
		dataChannel.OnOpen(func() {
			log.Printf("Data channel '%s'-'%dataChannel' open.\n", dataChannel.Label(), dataChannel.ID())

			// Notify sender we are ready to receive a file offer
			msg := Exchange{Type: "ready"}
			m, err := json.Marshal(msg)
			if err != nil {
				panic(err)
			}
			sendErr := dataChannel.Send(m)
			if sendErr != nil {
				panic(sendErr)
			}
		})

		// Register text message handling
		rebuiltFile := []byte{}

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {

			var m Exchange

			err := json.Unmarshal(msg.Data, &m)
			if err != nil {
				panic(err)
			}
			switch m.Type {

			// a file offer has been received
			case "fileInfo":
				// display received file
				fmt.Printf("Received a file offer:\nName: %s\nSize: %d byte\n", m.FileName, m.FileSize)
				var userResponse string
				fmt.Println("Type 'yes' to accept offer:")
				fmt.Scanln(&userResponse)
				switch userResponse {
				case "yes", "y":
					log.Println("downloading")

					outputPath = "out/" + m.FileName
					fileHash = m.Hash

					msg := Exchange{Type: "accept"}
					m, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					sendErr := dataChannel.Send(m)
					if sendErr != nil {
						panic(sendErr)
					}

				case "n", "no":
					msg := Exchange{Type: "reject"}
					m, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					sendErr := dataChannel.Send(m)
					if sendErr != nil {
						panic(sendErr)
					}

				default:
					fmt.Println("Unknown command, please try again:")
				}

			// when we receive a chunk of file, adds it to the file byte array
			case "fileChunk":
				rebuiltFile = append(rebuiltFile, m.Data[:]...)

			// each time one 1% of all chunks have been received
			case "mega":
				fmt.Print("|")

			// when all chunks of file have been received
			case "fileComplete":
				fmt.Println(" -->Download done")
				log.Println("Integrity check")
				// compares received filehash and generated one to confirm integrity
				if bytes.Equal(fileHash, internal.CreateHash(rebuiltFile)) {
					fmt.Println("File integrity confirmed! Saving file...")

					err = ioutil.WriteFile(outputPath, rebuiltFile, 0644)
					if err != nil {
						panic(err)
					}

					msg := Exchange{Type: "received"}
					m, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					sendErr := dataChannel.Send(m)
					if sendErr != nil {
						panic(sendErr)
					}
				} else {
					msg := Exchange{Type: "transferfailed"}
					m, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					sendErr := dataChannel.Send(m)
					if sendErr != nil {
						panic(sendErr)
					}
				}
			case "newfile":
				fmt.Println("Remote user wants to send you another file.")
				fmt.Println("Keep receiving files? ('y'/'n')")
				newfile := ""
				fmt.Scanln(&newfile)
				switch newfile {
				case "yes", "y":
					msg := Exchange{Type: "ready"}
					m, err := json.Marshal(msg)
					if err != nil {
						panic(err)
					}
					sendErr := dataChannel.Send(m)
					if sendErr != nil {
						panic(sendErr)
					}
				}

			}
		})
	})

	//////////////////////////////////////////// WEBSOCKET SIGNALING ///////////////////////////////////////////////////

	type Message struct {
		Type    string
		Success bool
		Offer   string
		Answer  string
		Name    string
		Sender  string
	}

	var remote string
	linked := false

	// ask user for remote passphrase
	fmt.Println("Enter your sender's passphrase:")
	fmt.Scanln(&remote)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New("ws://127.0.0.1:9090")
	//socket := gowebsocket.New("ws://signal.flying-dut.ch:9090")

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")

		ans := Message{Type: "login", Name: internal.Reverse(remote + localPassphrase)}
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
		var m Message

		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			panic(err)
		}
		switch m.Type {
		case "login":
			if m.Success == true {
				log.Println("Login success, searching for remote user...")

			} else {
				log.Println("Login failed")
			}
		case "linked":
			linked = true
			log.Println("Linked !")

		case "offer":
			log.Println("Received offer from " + m.Name)

			var encodedOffer = m.Offer
			offer := webrtc.SessionDescription{}
			internal.Decode(encodedOffer, &offer)

			// Checking remote certificate's fingerprint matches given passphrase
			parsed := &sdp.SessionDescription{}
			if err := parsed.Unmarshal([]byte(offer.SDP)); err != nil {
				panic(err)
			}
			fingerprint := internal.ExtractFingerprint(parsed)
			remotePassphrase := internal.FingerprintToPhrase(fingerprint)

			// If certificate matches, set as remote description
			if remotePassphrase == remote {
				fmt.Println("Receiver identity confirmed!")
				err = peerConnection.SetRemoteDescription(offer)
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println("Receiver's certificate is not matching")
				break
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
			<-gatherComplete

			encodedAnswer := internal.Encode(*peerConnection.LocalDescription())

			ans := Message{Type: "answer", Name: remote, Answer: encodedAnswer}
			b, err := json.Marshal(ans)
			if err != nil {
				panic(err)
			}
			socket.SendBinary(b)
			log.Println("Sending answer to " + remote)

		case "leave":
			socket.Close()
			return
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
}
