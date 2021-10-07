package main

import (
	"flag"
	"log"
	"net/http"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/logging"
	"mealswipe.app/mealswipe/internal/websockets"
)

var addr = flag.String("addr", ":8080", "http service address")

// TODO NULL SAFETY FROM PROTOBUF STUFF
func main() {
	logger, err := zap.NewProduction(zap.Fields(zap.String("app", "ms-ws")))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logging.SetLogger(logger)

	// Connect to redis
	business.LoadRedisClient()

	// Honestly not sure
	flag.Parse()
	log.SetFlags(0)

	// Start the websocket server
	logger.Info("init")
	http.HandleFunc("/", websockets.WebsocketHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
