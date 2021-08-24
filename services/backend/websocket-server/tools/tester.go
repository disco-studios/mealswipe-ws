// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
// var addr = flag.String("addr", "k8s-default-mealswip-1089bac565-1635200056.us-east-1.elb.amazonaws.com", "http service address")

var addr = flag.String("addr", "mealswipesessions.discostudios.io", "http service address")

var sockets []*websocket.Conn

func write_message(index int, message *mealswipepb.WebsocketMessage) {
	out, err := proto.Marshal(message)

	connection := sockets[index]

	err = connection.WriteMessage(websocket.BinaryMessage, out)
	if err != nil {
		log.Println("write:", err)
		return
	}
	log.Println("\t("+fmt.Sprint(index)+") wrote", message)
}

func write_message_delay(index int, message *mealswipepb.WebsocketMessage) {
	write_message(index, message)
	time.Sleep(time.Second * 3)
}

func spawn(lobbyCode *string, u url.URL) int {
	socketNum := len(sockets)
	// Connect to host websocket
	wsCon, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial"+fmt.Sprint(socketNum)+":", err)
	}

	go func() {
		for {
			_, messageBytes, err := wsCon.ReadMessage()
			if err != nil {
				log.Println("\t("+fmt.Sprint(socketNum)+") read error"+fmt.Sprint(socketNum)+":", err)
				return
			}
			genericMessage := &mealswipepb.WebsocketResponse{}
			if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
				log.Println("\t("+fmt.Sprint(socketNum)+") read decode error: ", err)
				return
			}
			if genericMessage.GetLobbyInfoMessage() != nil {
				(*lobbyCode) = genericMessage.GetLobbyInfoMessage().Code
				log.Println("\t("+fmt.Sprint(socketNum)+") got code:", *lobbyCode)
			}
			if genericMessage.GetGameStartedMessage() != nil {
				log.Println("\t(" + fmt.Sprint(socketNum) + ") Got game started!")
			}
			log.Println("\t("+fmt.Sprint(socketNum)+") received", genericMessage)
		}
	}()

	sockets = append(sockets, wsCon)
	return socketNum
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())
	var lobbyCode string

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n\n\n\n------------------------------------------------------------------------------------")
	fmt.Println("| Mealswipe tester                                                                 |")
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println("| * spawn                         | creates a new websocket instance               |")
	fmt.Println("| * create <ws#> <nickname>       | creates a lobby with ws <ws#>                  |")
	fmt.Println("| * join <ws#> <nickname> <code>  | joins ws <ws#> into lobby <code>               |")
	fmt.Println("| * start <ws#> <lat> <lng> <rad> | uses ws <ws#> to start at <lat>,<lng> in <rad> |")
	fmt.Println("|         <ws#>                   | > starts with a default location of philly     |")
	fmt.Println("| * vote <ws#> <ind> <y/n>        | votes for ws <ws#>                             |")
	fmt.Println("| * q                             | quits                                          |")
	fmt.Println("------------------------------------------------------------------------------------")

	for {
		fmt.Print("$ ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "\r", "", -1)
		parts := strings.Split(text, " ")

		if len(parts) == 0 || strings.Compare("", parts[0]) == 0 {
			fmt.Println("Please provide input.")
		} else if strings.Compare("spawn", parts[0]) == 0 {
			socketNum := spawn(&lobbyCode, u)
			if socketNum >= len(sockets) {
				fmt.Println("> That socket doesn't exist dummy")
				continue
			}
			fmt.Println("> Spawned ws#", socketNum)
		} else if strings.Compare("create", parts[0]) == 0 {
			if len(parts) != 3 {
				fmt.Println("> Please do better. Wrong arg count.")
			} else {
				socketNum, _ := strconv.Atoi(parts[1])
				if socketNum >= len(sockets) {
					fmt.Println("> That socket doesn't exist dummy")
					continue
				}
				write_message_delay(socketNum, &mealswipepb.WebsocketMessage{
					CreateMessage: &mealswipepb.CreateMessage{
						Nickname: parts[2],
					},
				})
				fmt.Println("> Done")
			}
		} else if strings.Compare("join", parts[0]) == 0 {
			if len(parts) != 4 {
				fmt.Println("> Please do better. Wrong arg count.")
			} else {
				socketNum, _ := strconv.Atoi(parts[1])
				if socketNum >= len(sockets) {
					fmt.Println("> That socket doesn't exist dummy")
					continue
				}
				write_message_delay(socketNum, &mealswipepb.WebsocketMessage{
					JoinMessage: &mealswipepb.JoinMessage{
						Nickname: parts[2],
						Code:     parts[3],
					},
				})
				fmt.Println("> Done")
			}
		} else if strings.Compare("start", parts[0]) == 0 {
			if len(parts) == 5 {
				socketNum, _ := strconv.Atoi(parts[1])
				if socketNum >= len(sockets) {
					fmt.Println("> That socket doesn't exist dummy")
					continue
				}
				lat, _ := strconv.ParseFloat(parts[2], 64)
				lng, _ := strconv.ParseFloat(parts[3], 64)
				rad, _ := strconv.Atoi(parts[4])

				write_message_delay(socketNum, &mealswipepb.WebsocketMessage{
					StartMessage: &mealswipepb.StartMessage{
						Lat:    lat,
						Lng:    lng,
						Radius: int32(rad),
					},
				})
				fmt.Println("> Done")
			} else if len(parts) == 2 {
				socketNum, _ := strconv.Atoi(parts[1])
				if socketNum >= len(sockets) {
					fmt.Println("> That socket doesn't exist dummy")
					continue
				}
				write_message_delay(socketNum, &mealswipepb.WebsocketMessage{
					StartMessage: &mealswipepb.StartMessage{
						Lat:    39.9533952,
						Lng:    -75.1882669,
						Radius: 1000,
					},
				})
				fmt.Println("> Done")
			} else {
				fmt.Println("> Please do better. Wrong arg count.")
			}
		} else if strings.Compare("vote", parts[0]) == 0 {
			if len(parts) != 4 {
				fmt.Println("> Please do better. Wrong arg count.")
			} else {
				socketNum, _ := strconv.Atoi(parts[1])
				if socketNum >= len(sockets) {
					fmt.Println("> That socket doesn't exist dummy")
					continue
				}
				ind, _ := strconv.Atoi(parts[2])
				write_message_delay(socketNum, &mealswipepb.WebsocketMessage{
					VoteMessage: &mealswipepb.VoteMessage{
						Index: int32(ind),
						Vote:  strings.Compare("t", parts[3]) == 0,
					},
				})
				fmt.Println("> Done")
			}
		}

		if strings.Compare("q", parts[0]) == 0 {
			fmt.Println("Goodbye!")
			// TODO Cleanup sockets
			os.Exit(0)
		}
	}

	// /*
	// *
	// *
	// * Create websocket connections
	// *
	// *
	//  */
	// wsCount := 2
	// for socketNum := 0; socketNum < wsCount; socketNum++ {
	// 	// Connect to host websocket
	// 	wsCon, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// 	if err != nil {
	// 		log.Fatal("dial"+fmt.Sprint(socketNum)+":", err)
	// 	}
	// 	defer wsCon.Close()

	// 	done := make(chan struct{})

	// 	go func() {
	// 		defer close(done)
	// 		for {
	// 			_, messageBytes, err := wsCon.ReadMessage()
	// 			if err != nil {
	// 				log.Println("read"+fmt.Sprint(socketNum)+":", err)
	// 				return
	// 			}
	// 			genericMessage := &mealswipepb.WebsocketResponse{}
	// 			if err := proto.Unmarshal(messageBytes, genericMessage); err != nil {
	// 				log.Println(fmt.Sprint(socketNum)+"read decode: ", err)
	// 				return
	// 			}
	// 			if genericMessage.GetLobbyInfoMessage() != nil {
	// 				lobbyCode = genericMessage.GetLobbyInfoMessage().Code
	// 				log.Println("("+fmt.Sprint(socketNum)+") Code:", lobbyCode)
	// 			}
	// 			if genericMessage.GetGameStartedMessage() != nil {
	// 				log.Println("(" + fmt.Sprint(socketNum) + ") Game started!")
	// 			}
	// 			log.Println(genericMessage)
	// 		}
	// 	}()

	// 	sockets = append(sockets, wsCon)
	// }

	// /*
	// *
	// *
	// * Send messages to websockets
	// *
	// *
	//  */

	// // User 0 creates lobby
	// write_message_delay(0, &mealswipepb.WebsocketMessage{
	// 	CreateMessage: &mealswipepb.CreateMessage{
	// 		Nickname: "Cam the Man",
	// 	},
	// })

	// // User 1 joins lobby
	// write_message_delay(1, &mealswipepb.WebsocketMessage{
	// 	JoinMessage: &mealswipepb.JoinMessage{
	// 		Nickname: "Bob the Builder",
	// 		Code:     lobbyCode,
	// 	},
	// })

	// // User 0 starts lobby
	// write_message_delay(0, &mealswipepb.WebsocketMessage{
	// 	StartMessage: &mealswipepb.StartMessage{
	// 		Lat:    39.9533952,
	// 		Lng:    -75.1882669,
	// 		Radius: 500,
	// 	},
	// })

	// // User 0 votes
	// write_message_delay(0, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 0,
	// 		Vote:  true,
	// 	},
	// })
	// write_message_delay(0, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 1,
	// 		Vote:  true,
	// 	},
	// })
	// write_message_delay(0, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 2,
	// 		Vote:  true,
	// 	},
	// })

	// // User 1 votes
	// write_message_delay(1, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 0,
	// 		Vote:  false,
	// 	},
	// })
	// write_message_delay(1, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 1,
	// 		Vote:  false,
	// 	},
	// })
	// write_message_delay(1, &mealswipepb.WebsocketMessage{
	// 	VoteMessage: &mealswipepb.VoteMessage{
	// 		Index: 2,
	// 		Vote:  true,
	// 	},
	// })

	// /*
	// *
	// *
	// * Clean up
	// *
	// *
	//  */
	// log.Println("cleaning up")

	// for _, wsCon := range sockets {
	// 	err := wsCon.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 	if err != nil {
	// 		log.Println("write close:", err)
	// 		return
	// 	}
	// }
}
