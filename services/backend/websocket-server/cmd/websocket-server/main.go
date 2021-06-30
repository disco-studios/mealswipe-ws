package main

import (
	"flag"
	"log"
	"net/http"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/websockets"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	// Connect to redis
	business.LoadRedisClient()

	// Honestly not sure
	flag.Parse()
	log.SetFlags(0)

	// Start the websocket server
	log.Println("server init")
	http.HandleFunc("/v2/api", websockets.WebsocketHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
