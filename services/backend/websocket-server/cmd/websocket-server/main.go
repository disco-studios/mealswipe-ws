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

	// Test code reservations and collision rate
	// for i := 0; i < 1000000; i++ {
	// 	_, err := core.CreateSession()
	// 	if err != nil {
	// 		log.Println("failed to make session", err)
	// 	} else {
	// 	}
	// }

	// Test grabbing and caching
	// log.Println("Grabbing locations")
	// venues, err := business.GrabLocations(39.9533952, -75.1882669)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Println("Tagging locations")
	// err = business.TagLocationGrabs(venues)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Println("Grabbing detail for ", venues[0].ID)
	// loc, err := business.GetDetailedLocation(venues[0].ID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println("Detailed name ", loc.Name)

	// Start the websocket server
	log.Println("server init")
	http.HandleFunc("/v2/api", websockets.WebsocketHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
