package core

import (
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"mealswipe.app/mealswipe/business"
	"mealswipe.app/mealswipe/mealswipepb"
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
	return business.PubsubWrite("session."+userState.JoinedSessionId, message)
}

func CreateUserState() *UserState {
	userState := &UserState{}
	userState.HostState = HostState_UNIDENTIFIED
	userState.UserId = "u-" + uuid.NewString()
	userState.PubsubChannel = make(chan string, 5)

	return userState
}
