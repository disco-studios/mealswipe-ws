package users

import (
	"encoding/json"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

type UserState struct {
	HostState         int16
	JoinedSessionId   string
	JoinedSessionCode string
	UserId            string
	Nickname          string
	WriteChannel      chan *mealswipepb.WebsocketResponse
	PubsubChannel     chan string
	RedisPubsub       *redis.PubSub
}

func (userState UserState) SendWebsocketMessage(message *mealswipepb.WebsocketResponse) {
	userState.WriteChannel <- message
}

func (userState UserState) SendPubsubMessage(message string) (err error) {
	if userState.JoinedSessionId == "" {
		return errors.New("not currently in a session")
	}
	return business.DbPubsubWrite("session."+userState.JoinedSessionId, message)
}

func (userState UserState) PubsubWebsocketResponse(websocketResponse *mealswipepb.WebsocketResponse) (err error) {
	var bytes []byte
	bytes, err = json.Marshal(websocketResponse)
	if err != nil {
		return
	}

	err = userState.SendPubsubMessage(string(bytes))
	if err != nil {
		return
	}
	return
}

func CreateUserState() *UserState {
	userState := &UserState{}
	userState.HostState = constants.HostState_UNIDENTIFIED
	userState.UserId = "u-" + uuid.NewString()
	userState.PubsubChannel = make(chan string, 5)

	return userState
}
