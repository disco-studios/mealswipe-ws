package main

import (
	"flag"
	"log"
	"net/http"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/msredis"
	"mealswipe.app/mealswipe/internal/websockets"
)

var addr = flag.String("addr", ":8080", "http service address")
var ctlAddr = flag.String("ctlAddr", ":8081", "control http service address")

func handlePreStop(w http.ResponseWriter, r *http.Request) {
	websockets.Decommission()
	logging.Get().Core().Sync() // Flush out logs
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text")
	w.Write([]byte("Done"))
}

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
	msredis.LoadRedisClient()

	// Honestly not sure
	flag.Parse()

	// Handle kubernetes hooks
	kubehooks := http.NewServeMux()
	kubehooks.HandleFunc("/preStop", handlePreStop)
	kubehookserver := http.Server{
		Addr:    *ctlAddr,
		Handler: kubehooks,
	}
	go func() {
		err := kubehookserver.ListenAndServe()
		logger.Error("kube hook server failed", zap.Error(err))
		websockets.Decommission()
		logging.Get().Core().Sync()
		logger.Fatal("kube hook server failed", zap.Error(err))
	}()

	// Start the websocket server
	logger.Info("init")
	http.HandleFunc("/", websockets.WebsocketHandler)

	err = http.ListenAndServe(*addr, nil)
	logger.Error("http server failed", zap.Error(err))
	websockets.Decommission()
	logging.Get().Core().Sync()
	logger.Fatal("http server failed", zap.Error(err))
}
