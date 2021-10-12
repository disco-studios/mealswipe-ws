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
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "@timestamp"
	loggerConfig.EncoderConfig.MessageKey = "message"
	logger, err := loggerConfig.Build(zap.Fields(zap.String("app", "ms-ws")))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logging.SetLogger(logger)

	// Connect to redis
	business.LoadRedisClient()

	// Honestly not sure
	flag.Parse()

	// Start the websocket server
	logger.Info("init")
	http.HandleFunc("/", websockets.WebsocketHandler)
	logger.Fatal("http server failed", zap.Error(http.ListenAndServe(*addr, nil)))
}
