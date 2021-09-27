package main

import (
	"flag"
	"log"
	"net/http"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/websockets"
)

var addr = flag.String("addr", ":8080", "http service address")

// TODO NULL SAFETY FROM PROTOBUF STUFF
func main() {
	// Connect to redis
	business.LoadRedisClient()

	// Honestly not sure
	flag.Parse()
	log.SetFlags(0)

	// Start the websocket server
	log.Println("server init")
	http.HandleFunc("/", websockets.WebsocketHandler) // /v2/api
	log.Fatal(http.ListenAndServe(*addr, nil))
}
