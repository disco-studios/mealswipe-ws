package websockets

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

/*
This is the main file for handling websocket connections.
For each connection, the following is made:

- readPump
- writePump
- pubsubPump

These are used to handle messages for the various channels we communicate over. We can only
read/write one message at a time, so these handlers deal with that.

TODO: Ping/pong or read/write deadlines, buffering, TLS (working from client to ALB, but verify that works, and see if we need tls from alb to server)
*/

var websocketUpgrader = websocket.Upgrader{} // use default options
var localSessions types.LocalSessions = *types.InitLocalSessions()

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.Get()
	// Upgrade the connection to a websocket
	c, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Debug("ws failed to upgrade connection", zap.Error(err))
		return
	}
	defer c.Close()

	// Create a user state for our user
	userState := types.CreateUserState()
	localSessions.Add(userState)
	defer localSessions.Remove(userState)
	defer ensureCleanup(userState)
	defer close(userState.PubsubChannel)
	go pubsubPump(userState, userState.PubsubChannel)
	logger.Info("new user connected", logging.UserId(userState.UserId))

	// Create a write channel to send messages to our websocket
	writeChannel := make(chan *mealswipepb.WebsocketResponse, 5)
	defer close(writeChannel)
	userState.WriteChannel = writeChannel

	// Start a write pump to watch our channel and send messages when we need to
	go writePump(c, writeChannel)

	// Call the read pump to handle incoming messages
	readPump(c, userState)
}
