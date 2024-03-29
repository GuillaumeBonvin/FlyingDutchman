package main

import (
	"FlyingDutchman/internal"
	"encoding/json"
	"fmt"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v3"
	"github.com/sacOO7/gowebsocket"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

func Sender() {

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:       []string{"turn:turn.flying-dut.ch:3478", "stun:stun.flying-dut.ch:3478"},
				Username:   "captain",
				Credential: "Axp2oSr56d5"},
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

	// Create a datachannel with label 'data'
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	////////////////////////////////////////// FILE EXCHANGE PROTOCOL //////////////////////////////////////////////////

	type Exchange struct {
		Type     string
		FileName string
		FileSize int64
		Hash     []byte
		Data     []byte
	}

	var filePath string
	var file []byte
	var fileStats os.FileInfo

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		log.Printf("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())

	})

	// Register text message handling
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		//fmt.Printf("Message from DataChannel '%s': '%s'\n", dataChannel.Label(), string(msg.Data))

		var m Exchange

		err := json.Unmarshal(msg.Data, &m)
		if err != nil {
			panic(err)
		}
		switch m.Type {

		// receiver sends confirmation connection is ready, waiting for a file transfer offer
		case "ready":
			fmt.Println("Receiver is ready for a file offer, please enter file path and name:\n" +
				"Example - somefolder/image.png")
			var userInput string
			fmt.Scanln(&userInput)
			if userInput == "q" { //send leave message
			} else {
				filePath = userInput
				fileStats, err = os.Stat(filePath)
				if err != nil {
					log.Fatal("err")
				}
				// Convert file to byte array
				file, err = ioutil.ReadFile(filePath)
				if err != nil {
					panic(err)
				}
				fileHash := internal.CreateHash(file)

				// Send all file infos
				msg := Exchange{Type: "fileInfo", FileName: fileStats.Name(), FileSize: fileStats.Size(), Hash: fileHash}
				m, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}
				sendErr := dataChannel.Send(m)
				if sendErr != nil {
					panic(sendErr)
				}
			}

			// receiver has accepted our file offer, upload can be done
		case "accept":
			fmt.Println("File offer accepted! Your file is being sent...")
			log.Println("Uploading")

			// sends selected file chunk by chunk
			limit := 45000
			for i := 0; i < len(file); i += limit {
				batch := file[i:internal.Min(i+limit, len(file))]

				msg := Exchange{Type: "fileChunk", Data: batch}
				m, err := json.Marshal(msg)
				if err != nil {
					panic(err)
				}
				sendErr := dataChannel.Send(m)
				if sendErr != nil {
					panic(sendErr)
				}

				chunks := len(file) / limit
				if chunks < 100 {
					chunks = 100
				}

				// each 1%, send a notification message
				if (i/limit)%(chunks/100) == 0 && i != 0 {
					fmt.Print("|")
					msg := Exchange{Type: "mega"}
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
			fmt.Println(" -->Upload done\nWaiting for confirmation...")
			msg := Exchange{Type: "fileComplete"}
			m, err := json.Marshal(msg)
			if err != nil {
				panic(err)
			}
			sendErr := dataChannel.Send(m)
			if sendErr != nil {
				panic(sendErr)
			}

			// file has been successfully received and its integrity has been confirmed
		case "received":
			fmt.Println("File has been received successfully!")

			time.Sleep(50000 * time.Microsecond)
			peerConnection.Close()

			time.Sleep(50000 * time.Microsecond)
			main()
		}

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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// define websocket connection to signaling server
	socket := gowebsocket.New("ws://127.0.0.1:9090")
	//socket := gowebsocket.New("ws://signal.flying-dut.ch:9090")

	// on connection: send login info to signaling server
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")

		ans := Message{Type: "login", Name: localPassphrase}
		b, err := json.Marshal(ans)
		if err != nil {
			panic(err)
		}
		socket.SendBinary(b)
	}

	// on error: dislay error message
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Recieved connect error ", err)
	}

	// on text message: read its content and switch between cases
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		var m Message

		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			panic(err)
		}
		switch m.Type {

		// signaling server returns its login answer
		case "login":
			// login was successful
			if m.Success == true {
				log.Println("Login success")

				// create a new peerConnection offer
				offer, err := peerConnection.CreateOffer(nil)
				// gather candidates
				gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

				err = peerConnection.SetLocalDescription(offer)
				if err != nil {
					panic(err)
				}
				<-gatherComplete

				// output the answer in base64 so we can send it
				encodedOffer := internal.Encode(*peerConnection.LocalDescription())

				// ask user for remote passphrase
				fmt.Println("Enter your receiver's passphrase:")
				fmt.Scanln(&remote)

				// send offer to remote user connected with given passphrase
				ans := Message{Type: "offer", Name: remote, Offer: encodedOffer, Sender: localPassphrase}
				b, err := json.Marshal(ans)
				if err != nil {
					panic(err)
				}
				socket.SendBinary(b)

				log.Println("Sending offer to " + remote)
			} else {
				log.Println("Login failed")
			}

			// remote user couldn't be found or rejected our offer
		case "noMatch", "reject":

			if m.Type == "noMatch" {
				fmt.Println("Couldn't find user named " + remote)
			} else {
				fmt.Println("User " + remote + " rejected you offer")
			}

			// ask user for a new passphrase and repeat offer process
			fmt.Println("Please enter new name or type 'r' to retry")
			var userInput string
			fmt.Scanln(&userInput)
			if userInput != "r" {
				remote = userInput
			}

			offer, err := peerConnection.CreateOffer(nil)

			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

			err = peerConnection.SetLocalDescription(offer)
			if err != nil {
				panic(err)
			}
			<-gatherComplete

			// Output the answer in base64
			encodedOffer := internal.Encode(*peerConnection.LocalDescription())

			ans := Message{Type: "offer", Name: remote, Offer: encodedOffer, Sender: localPassphrase}
			b, err := json.Marshal(ans)
			if err != nil {
				panic(err)
			}
			socket.SendBinary(b)

			log.Println("Sending offer to " + remote)

			// remote user accepted our offer and sent us its answer
		case "answer":
			log.Println("Received answer from " + remote)
			var encodedAnswer = m.Answer

			answer := webrtc.SessionDescription{}

			internal.Decode(encodedAnswer, &answer)

			// Checking remote certificate's fingerprint matches given passphrase
			parsed := &sdp.SessionDescription{}
			if err := parsed.Unmarshal([]byte(answer.SDP)); err != nil {
				panic(err)
			}
			fingerprint := internal.ExtractFingerprint(parsed)
			remotePassphrase := internal.FingerprintToPhrase(fingerprint)

			// If certificate matches, set as remote description
			if remotePassphrase == remote {
				fmt.Println("Receiver identity confirmed!")
				err = peerConnection.SetRemoteDescription(answer)
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println("Receiver's certificate is not matching")
				break
			}

		case "leave":
			/*ans := Message{Type: "leave", Name: remote}
			b, err := json.Marshal(ans)
			if err != nil {
				panic(err)
			}
			socket.SendBinary(b)*/

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

	// connect to websocket connection and wait for event
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
