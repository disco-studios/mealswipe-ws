package websockets

import (
	"bytes"
	"encoding/json"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/messages"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// Clean up things we couldn't directly defer, because they are defined in different scopes
func ensureCleanup(userState *types.UserState) {
	if userState.RedisPubsub != nil {
		defer userState.RedisPubsub.Close()
	}
}

func pubsubPump(userState *types.UserState, messageQueue <-chan string) {
	logger := logging.Get()
	for message := range messageQueue {
		logger.Debug("redis message received", logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId), zap.String("message", message))

		websocketResponse := &mealswipepb.WebsocketResponse{}
		err := json.Unmarshal([]byte(message), websocketResponse) // We use the above as JSON, because we can, and it is easier to stringify
		if err != nil {
			logger.Error("pubsump pump failed to unmarshal protobuf", zap.Error(err))
		} else {
			userState.SendWebsocketMessage(websocketResponse)
			if websocketResponse.GetGameStartedMessage() != nil {
				for i := 0; i < 2; i++ {
					err := sessions.SendNextLocToUser(userState)
					if err != nil {
						logger.Error("pumpsump pump failed to send next location to user", zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
					}
				}
			}
		}
	}
	logger.Debug("pubsub pump cleaned up", logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
}

func writePump(connection *websocket.Conn, messageQueue <-chan *mealswipepb.WebsocketResponse) {
	logger := logging.Get()
	// Go through messages in the write pump and write them out
	for message := range messageQueue {
		w, err := connection.NextWriter(websocket.BinaryMessage) // TODO Does this need to happen in here? Feel like this can go out of loop
		if err != nil {
			logger.Error("write pump failed to open writer", zap.Error(err))
			return
		}

		out, err := proto.Marshal(message)
		if err != nil {
			logger.Error("write pump failed to marshal message to proto", zap.Error(err))
			return
		}

		outLength, err := w.Write(out)
		if err != nil {
			logger.Error("write pump failed to write message out", zap.Error(err))
			return
		}
		logger.Info("out message length", logging.Metric("out_message_length"), zap.Int("length", outLength))

		if err := w.Close(); err != nil {
			logger.Error("write pump failed on close", zap.Error(err))
			return
		}
	}
	logger.Debug("write pump cleaned up")
}

func readPump(connection *websocket.Conn, userState *types.UserState) {
	logger := logging.Get()
	for {
		// Establish read connection
		rawMessageType, inStream, err := connection.NextReader()
		if err != nil {
			logger.Error("read pump failed to open reader", zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
			return
		}
		if rawMessageType != websocket.BinaryMessage {
			logger.Info("user provided non-binary message", logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
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
			logger.Error("read pump failed when reading message", zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
			return
		}
		logger.Info("message length", logging.Metric("in_message_length"), zap.Int64("length", readLength), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
		messageBytes := readBuffer.Bytes()

		// Convert to generic message
		genericMessage := &mealswipepb.WebsocketMessage{}
		if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
			logger.Error("read pump failed to unmarshal protobuf", zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
			return
		}

		err = messages.ValidateMessage(userState, genericMessage)
		if err != nil {
			logger.Info("failed to validate message", logging.Metric("in_message_length"), zap.Any("raw", genericMessage), zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId))
			return
		}

		err = messages.HandleMessage(userState, genericMessage)
		if err != nil {
			// TODO Don't always die when we have an error, just sometimes
			logger.Error("message handler encountered error", zap.Error(err), logging.UserId(userState.UserId), logging.SessionId(userState.JoinedSessionId), zap.Any("raw", genericMessage))
			return
		}

		// TODO Close socket cleanly when we fail
	}
}
