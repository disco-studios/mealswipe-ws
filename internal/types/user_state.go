package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/msredis"
	"mealswipe.app/mealswipe/pkg/mealswipe"
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
		return errors.New("user not currently in a session")
	}
	return msredis.PubsubWrite(context.TODO(), keys.BuildSessionKey(userState.JoinedSessionId, ""), message)
}

func (userState UserState) PubsubWebsocketResponse(websocketResponse *mealswipepb.WebsocketResponse) (err error) {
	var bytes []byte
	bytes, err = json.Marshal(websocketResponse)
	if err != nil {
		err = fmt.Errorf("user marshal message for pubsub: %w", err)
		return
	}

	err = userState.SendPubsubMessage(string(bytes))
	if err != nil {
		err = fmt.Errorf("send message on pubsub: %w", err)
		return
	}
	return
}

func (userState UserState) TagContext(ctx context.Context) (newCtx context.Context) {
	newCtx = context.WithValue(ctx, "host.state", userState.HostState)
	if userState.UserId != "" {
		newCtx = context.WithValue(newCtx, "user.id", userState.UserId)
	}
	if userState.JoinedSessionId != "" {
		newCtx = context.WithValue(newCtx, "session.id", userState.JoinedSessionId)
	}
	return newCtx
}

func CreateUserState() *UserState {
	userState := &UserState{}
	userState.HostState = mealswipe.HostState_UNIDENTIFIED
	userState.UserId = "u-" + uuid.NewString()
	userState.PubsubChannel = make(chan string, 5)

	return userState
}

// We want to keep track of all sessions, but we are logging them from many threads
// Create this to keep track of running sessions without much complexity in implementation
// TODO Create a metric on this
type LocalSessions struct {
	sync.RWMutex
	internal map[string]*UserState
}

func InitLocalSessions() *LocalSessions {
	return &LocalSessions{
		internal: make(map[string]*UserState),
	}
}

func (s *LocalSessions) Add(value *UserState) {
	s.Lock()
	s.internal[value.UserId] = value
	s.Unlock()
}

func (s *LocalSessions) Remove(value *UserState) {
	s.Lock()
	delete(s.internal, value.UserId)
	s.Unlock()
}

func (s *LocalSessions) GetAll() []*UserState {
	s.RLock()
	v := make([]*UserState, 0, len(s.internal))

	for _, value := range s.internal {
		v = append(v, value)
	}
	s.RUnlock()

	return v
}
