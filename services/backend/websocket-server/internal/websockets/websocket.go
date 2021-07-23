package websockets

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/internal/core/locations"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/internal/handlers"
	"mealswipe.app/mealswipe/internal/validators"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var websocketUpgrader = websocket.Upgrader{} // use default options

// Clean up things we couldn't directly defer, because they are defined in different scopes
func ensureCleanup(userState *users.UserState) {
	if userState.RedisPubsub != nil {
		defer userState.RedisPubsub.Close()
	}
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a websocket
	c, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO Disabled... error connection? Maybe re-enable
		// log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Create a user state for our user
	userState := users.CreateUserState()
	defer ensureCleanup(userState)
	defer close(userState.PubsubChannel)
	go pubsubPump(userState, userState.PubsubChannel)
	log.Println("New user " + userState.UserId)

	// Create a write channel to send messages to our websocket
	writeChannel := make(chan *mealswipepb.WebsocketResponse, 5)
	defer close(writeChannel)
	userState.WriteChannel = writeChannel

	// Start a write pump to watch our channel and send messages when we need to
	go writePump(c, writeChannel)

	// Call the read pump to handle incoming messages
	readPump(c, userState)
}

func pubsubPump(userState *users.UserState, messageQueue <-chan string) {
	for message := range messageQueue {
		log.Println("Redis message -> '" + message + "'")

		websocketResponse := &mealswipepb.WebsocketResponse{}
		err := json.Unmarshal([]byte(message), websocketResponse)
		if err != nil {
			log.Println("pubsub pump:", err)
		} else {
			userState.SendWebsocketMessage(websocketResponse)
			if websocketResponse.GetGameStartedMessage() != nil {
				err := locations.SendNextToUser(userState)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	log.Println("Pubsub cleaned up")
}

func writePump(connection *websocket.Conn, messageQueue <-chan *mealswipepb.WebsocketResponse) {
	// Go through messages in the write pump and write them out
	for message := range messageQueue {
		w, err := connection.NextWriter(websocket.BinaryMessage)
		if err != nil {
			log.Println("write open:", err)
			return
		}

		out, err := proto.Marshal(message)
		if err != nil {
			log.Println("write marshal:", err)
			return
		}

		_, err = w.Write(out)
		if err != nil {
			log.Println("write:", err)
			return
		}

		if err := w.Close(); err != nil {
			log.Println("write close:", err)
			return
		}
	}
	log.Println("WritePump cleaned up")
}

func readPump(connection *websocket.Conn, userState *users.UserState) {
	for {
		// Establish read connection
		rawMessageType, inStream, err := connection.NextReader()
		if err != nil {
			// Something went wrong
			log.Println("read:", err)
			return
		}
		if rawMessageType != websocket.BinaryMessage {
			log.Println("Provided non-binary message")
			return
		}

		// Send an ack so we know the message was received for debugging
		userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
			Ack: "ack",
		}) // TODO Remove for prod

		// Read in the raw message from the stream
		readBuffer := new(bytes.Buffer)
		readLength, readErr := readBuffer.ReadFrom(inStream)
		if readErr != nil {
			// Something went wrong
			log.Println("destream:", err)
			return
		}
		log.Println("Message length", readLength)
		messageBytes := readBuffer.Bytes()

		// Convert to generic message
		genericMessage := &mealswipepb.WebsocketMessage{}
		if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
			log.Println("decode: ", err)
			return
		}

		err = validators.ValidateMessage(userState, genericMessage)
		if err != nil {
			// TODO Don't always die when we have an error, just sometimes
			log.Println("validate: ", err)
			return
		}

		err = handlers.HandleMessage(userState, genericMessage)
		if err != nil {
			// TODO Don't always die when we have an error, just sometimes
			log.Println("handle: ", err)
			return
		}

		// TODO Close socket cleanly when we fail
	}
}
