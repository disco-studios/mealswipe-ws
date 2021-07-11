// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var addr = flag.String("addr", "k8s-default-mealswip-1089bac565-1035559386.us-east-1.elb.amazonaws.com", "http service address")

var sockets []*websocket.Conn

func write_message(index int, message *mealswipepb.WebsocketMessage) {
	out, err := proto.Marshal(message)

	connection := sockets[index]

	err = connection.WriteMessage(websocket.BinaryMessage, out)
	if err != nil {
		log.Println("write:", err)
		return
	}
	log.Println("Wrote")
}

func write_message_delay(index int, message *mealswipepb.WebsocketMessage) {
	write_message(index, message)
	time.Sleep(time.Second * 3)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())
	var lobbyCode string

	/*
	*
	*
	* Create websocket connections
	*
	*
	 */
	wsCount := 2
	for socketNum := 0; socketNum < wsCount; socketNum++ {
		// Connect to host websocket
		wsCon, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial"+fmt.Sprint(socketNum)+":", err)
		}
		defer wsCon.Close()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				_, messageBytes, err := wsCon.ReadMessage()
				if err != nil {
					log.Println("read"+fmt.Sprint(socketNum)+":", err)
					return
				}
				genericMessage := &mealswipepb.WebsocketResponse{}
				if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
					log.Println(fmt.Sprint(socketNum)+"read decode: ", err)
					return
				}
				if genericMessage.GetLobbyInfoMessage() != nil {
					lobbyCode = genericMessage.GetLobbyInfoMessage().Code
					log.Println("("+fmt.Sprint(socketNum)+") Code:", lobbyCode)
				}
				if genericMessage.GetGameStartedMessage() != nil {
					log.Println("(" + fmt.Sprint(socketNum) + ") Game started!")
				}
				log.Println(genericMessage)
			}
		}()

		sockets = append(sockets, wsCon)
	}

	/*
	*
	*
	* Send messages to websockets
	*
	*
	 */

	// User 0 creates lobby
	write_message_delay(0, &mealswipepb.WebsocketMessage{
		CreateMessage: &mealswipepb.CreateMessage{
			Nickname: "Cam the Man",
		},
	})

	// User 1 joins lobby
	write_message_delay(1, &mealswipepb.WebsocketMessage{
		JoinMessage: &mealswipepb.JoinMessage{
			Nickname: "Bob the Builder",
			Code:     lobbyCode,
		},
	})

	// User 0 starts lobby
	write_message_delay(0, &mealswipepb.WebsocketMessage{
		StartMessage: &mealswipepb.StartMessage{
			Lat: 39.9533952,
			Lng: -75.1882669,
		},
	})

	// User 0 votes
	write_message_delay(0, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 0,
			Vote:  true,
		},
	})
	write_message_delay(0, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 1,
			Vote:  true,
		},
	})
	write_message_delay(0, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 2,
			Vote:  true,
		},
	})

	// User 1 votes
	write_message_delay(1, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 0,
			Vote:  false,
		},
	})
	write_message_delay(1, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 1,
			Vote:  false,
		},
	})
	write_message_delay(1, &mealswipepb.WebsocketMessage{
		VoteMessage: &mealswipepb.VoteMessage{
			Index: 2,
			Vote:  true,
		},
	})

	/*
	*
	*
	* Clean up
	*
	*
	 */
	log.Println("cleaning up")

	for _, wsCon := range sockets {
		err := wsCon.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
			return
		}
	}
}
