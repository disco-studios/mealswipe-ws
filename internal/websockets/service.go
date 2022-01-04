package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/messages"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
)

// Clean up things we couldn't directly defer, because they are defined in different scopes
func ensureCleanup(userState *types.UserState) {
	if userState.RedisPubsub != nil {
		defer userState.RedisPubsub.Close()
	}
}

func pubsubPump(userState *types.UserState, messageQueue <-chan string) {
	for message := range messageQueue {
		ctx := userState.TagContext(context.Background())
		logger := logging.Ctx(ctx)

		logger.Debug("redis message received", zap.String("message", message))

		websocketResponse := &mealswipepb.WebsocketResponse{}
		err := json.Unmarshal([]byte(message), websocketResponse) // We use the above as JSON, because we can, and it is easier to stringify
		if err != nil {
			logger.Error("pubsump pump failed to unmarshal protobuf", zap.Error(err))
		} else {
			// Inject user information to lobby info message if present
			if common.HasLobbyInfoMessage(websocketResponse) {
				message := websocketResponse.GetLobbyInfoMessage()
				message.SessionId = userState.JoinedSessionId
				message.UserId = userState.UserId
			}

			// Send the response
			userState.SendWebsocketMessage(websocketResponse)

			// If it was a game started, send the next locs too
			if websocketResponse.GetGameStartedMessage() != nil {
				for i := 0; i < 3; i++ {
					err := sessions.SendNextLocToUser(context.TODO(), userState)
					if err != nil {
						logger.Error("pumpsump pump failed to send next location to user", zap.Error(err))
					}
				}
			}
		}
	}
	logging.Get().Debug("pubsub pump cleaned up")
}

func writePump(userState *types.UserState, connection *websocket.Conn, messageQueue <-chan *mealswipepb.WebsocketResponse) {
	// Go through messages in the write pump and write them out
	for message := range messageQueue {
		ctx := userState.TagContext(context.Background())
		logger := logging.Ctx(ctx)

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
		logging.MetricCtx(ctx, "out_message_length").Debug(
			fmt.Sprintf("out message length %d", outLength),
			zap.Int("length", outLength),
		)

		if err := w.Close(); err != nil {
			logger.Error("write pump failed on close", zap.Error(err))
			return
		}
	}
	logging.Get().Debug("write pump cleaned up")
}

func readPump(userState *types.UserState, connection *websocket.Conn) {
	for {
		ctx := userState.TagContext(context.Background())
		logger := logging.Ctx(ctx)

		// Establish read connection
		rawMessageType, inStream, err := connection.NextReader()
		if err != nil {
			logger.Error("read pump failed to open reader", zap.Error(err))
			return
		}
		if rawMessageType != websocket.BinaryMessage {
			logger.Info(fmt.Sprintf("user %s provided non-binary message", userState.UserId))
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
			logger.Error("read pump failed when reading message", zap.Error(err))
			return
		}
		logging.MetricCtx(ctx, "in_message_length").Debug(
			fmt.Sprintf("in message length %d", readLength),
			zap.Int64("length", readLength),
		)
		messageBytes := readBuffer.Bytes()

		// Convert to generic message
		genericMessage := &mealswipepb.WebsocketMessage{}
		if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
			logger.Error("read pump failed to unmarshal protobuf", zap.Error(err))
			return
		}

		// tx := apm.DefaultTracer.StartTransaction("HANDLE create", "request")
		// defer tx.End()
		// ctx = apm.ContextWithTransaction(ctx, tx)

		err, ws_err := messages.ValidateMessage(userState, genericMessage)
		if err != nil || ws_err != nil {
			logging.MetricCtx(ctx, "message_validation_failed").Info(
				"failed to validate message",
				zap.Any("raw", genericMessage),
				zap.Error(err),
			)
			if ws_err != nil {
				userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
					ErrorMessage: ws_err,
				})
			} else {
				userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
					ErrorMessage: &mealswipepb.ErrorMessage{
						ErrorType: mealswipepb.ErrorType_UnhandledError,
						Message:   "Unhandled validation error",
					},
				})
			}
		}

		err = messages.HandleMessage(ctx, userState, genericMessage)
		if err != nil {
			// TODO Don't always die when we have an error, just sometimes
			logging.Ctx(ctx).Error("message handler encountered error", zap.Error(err), zap.Any("raw", genericMessage))
			userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
				ErrorMessage: &mealswipepb.ErrorMessage{
					ErrorType: mealswipepb.ErrorType_UnhandledError,
					Message:   "Unhandled logic error",
				},
			})
		}

		// TODO Close socket cleanly when we fail
	}
}
