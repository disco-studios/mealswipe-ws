package sessions

import (
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/codes"
	"mealswipe.app/mealswipe/internal/locations"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GetIdFromCode(code string) (sessionId string, err error) {
	return getIdFromCode(code)
}

func GetActiveUsers(sessionId string) (activeUsers []string, err error) {
	return getActiveUsers(sessionId)
}

func GetActiveNicknames(sessionId string) (activeNicknames []string, err error) {
	return getActiveNicknames(sessionId)
}

func JoinById(userState *types.UserState, sessionId string, code string) (err error) {
	redisPubsub, err := joinById(userState.UserId, sessionId, userState.Nickname, userState.PubsubChannel)
	if err != nil {
		return
	}

	userState.RedisPubsub = redisPubsub
	userState.JoinedSessionId = sessionId
	userState.JoinedSessionCode = code

	return
}

// TODO This shouldn't live here
func HandleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
}

func Start(code string, sessionId string, lat float64, lng float64, radius int32, categoryId string) (err error) {

	// TODO This write should maybe go into locs
	venueIds, distances, err := locations.IdsForLocation(lat, lng, radius, categoryId)
	if err != nil {
		return
	}
	if len(venueIds) == 0 {
		return errors.New("found no venues for loc")
	}

	return start(code, sessionId, venueIds, distances)
}

func Vote(userId string, sessionId string, index int32, state bool) (err error) {
	return vote(userId, sessionId, index, state)
}

func CheckWin(userState *types.UserState) (err error) {
	win, winIndex, err := getWinIndex(userState.JoinedSessionId)
	if err != nil {
		return
	}

	if win {
		// TODO This shouldn't live here I don't think
		var loc *mealswipepb.Location
		loc, err = locations.FromInd(userState.JoinedSessionId, winIndex)
		if err != nil {
			return
		}

		err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
			GameWinMessage: &mealswipepb.GameWinMessage{
				Locations: []*mealswipepb.WinningLocation{
					{
						Location: loc,
						Votes:    0, // TODO: Impl
					},
				},
			},
		})
		if err != nil {
			return
		}
	}
	return
}

func Create(userState *types.UserState) (sessionID string, code string, err error) {
	sessionID = "s-" + uuid.NewString()
	code, err = codes.Reserve(sessionID)
	if err != nil {
		return
	}
	err = create(code, sessionID, userState.UserId)
	return
}

func GetNextLocForUser(userState *types.UserState) (loc *mealswipepb.Location, err error) {
	ind, err := nextVoteInd(userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, err = locations.FromInd(userState.JoinedSessionId, int32(ind))
	return
}

func SendNextLocToUser(userState *types.UserState) (err error) {
	loc, err := GetNextLocForUser(userState)
	if err != nil {
		return
	}

	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		Location: loc,
	})
	return
}