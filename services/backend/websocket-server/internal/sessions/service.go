package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/codes"
	"mealswipe.app/mealswipe/internal/common/logging"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/locations"
	"mealswipe.app/mealswipe/internal/msredis"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GetIdFromCode(code string) (sessionId string, err error) {
	key := keys.BuildCodeKey(code)
	result := msredis.GetRedisClient().Get(context.TODO(), key)
	return result.Val(), result.Err()
}

func GetActiveUsers(sessionId string) (activeUsers []string, err error) {
	hGetAll := msredis.GetRedisClient().HGetAll(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_ACTIVE))
	if err = hGetAll.Err(); err != nil {
		return
	}

	for userId, active := range hGetAll.Val() {
		if active == "1" {
			activeUsers = append(activeUsers, userId)
		}
	}

	if len(activeUsers) == 0 {
		err = errors.New("session has no active users")
		return
	}

	return
}

func GetActiveNicknames(sessionId string) (activeNicknames []string, err error) {
	activeUsers, err := GetActiveUsers(sessionId)
	if err != nil {
		return
	}

	hGetAll := msredis.GetRedisClient().HGetAll(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_NICKNAMES))
	if err = hGetAll.Err(); err != nil {
		return
	}

	nicknamesMap := hGetAll.Val()
	if err != nil {
		return
	}

	for _, userId := range activeUsers {
		activeNicknames = append(activeNicknames, nicknamesMap[userId])
	}

	return
}

func JoinById(userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	pipe := msredis.GetRedisClient().Pipeline()
	timeToLive := time.Hour * 24

	pipe.SetNX(context.TODO(), keys.BuildUserKey(userId, keys.KEY_USER_SESSION), sessionId, time.Hour*24)
	pipe.SAdd(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS), userId)
	pipe.HSet(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_ACTIVE), userId, true)        // TODO Expire
	pipe.HSet(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND), userId, 0)                // TODO Expire
	pipe.HSet(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_NICKNAMES), userId, nickname) // TODO Expire
	pipe.SetBit(context.TODO(), keys.BuildVotesKey(sessionId, userId), 0, 0)
	pipe.Expire(context.TODO(), keys.BuildVotesKey(sessionId, userId), timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	// Initiate a pubsub with this session
	redisPubsub = msredis.GetRedisClient().Subscribe(context.TODO(), keys.BuildSessionKey(sessionId, ""))
	pubsubChannel := redisPubsub.Channel()
	go HandleRedisMessages(pubsubChannel, genericPubsub)

	return
}

func HandleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
	logging.Get().Debug("redis pubsub cleaned up") // TODO Session/user id here would be good
}

func reverse(venues []string) []string {
	for i, j := 0, len(venues)-1; i < j; i, j = i+1, j-1 {
		venues[i], venues[j] = venues[j], venues[i]
	}
	return venues
}

func Start(code string, sessionId string, lat float64, lng float64, radius int32, categoryId string) (err error) {
	pipe := msredis.GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

	venueIds, distances, err := locations.IdsForLocation(lat, lng, radius, categoryId)
	if err != nil {
		return
	}
	if len(venueIds) == 0 {
		return errors.New("found no venues for loc")
	}

	pipe.Del(context.TODO(), keys.BuildCodeKey(code))
	pipe.Set(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_GAME_STATE), "RUNNING", timeToLive) // TODO pull RUNNING into constant
	pipe.LPush(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), reverse(venueIds))

	var distanceStrings []string
	for _, distance := range distances {
		distanceStrings = append(distanceStrings, fmt.Sprintf("%d", int(distance)))
	}

	pipe.LPush(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), reverse(distanceStrings))
	pipe.Expire(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), timeToLive)
	pipe.Expire(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	return
}

func Vote(userId string, sessionId string, index int32, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	return msredis.GetRedisClient().SetBit(context.TODO(), keys.BuildVotesKey(sessionId, userId), int64(index), voteBit).Err()
}

func dbGameCheckWin(sessionId string) (win bool, winningIndex int32, err error) {
	activeUsers, err := GetActiveUsers(sessionId)
	if err != nil {
		return
	}

	var voteKeys []string
	for _, userId := range activeUsers {
		voteKeys = append(voteKeys, keys.BuildVotesKey(sessionId, userId))
	}

	pipe := msredis.GetRedisClient().Pipeline()

	pipe.BitOpAnd(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), voteKeys...)
	winningIndexResult := pipe.BitPos(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), 1)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	winningIndex = int32(winningIndexResult.Val())
	win = winningIndex > -1
	return
}

func CheckWin(userState *users.UserState) (err error) {
	win, winIndex, err := dbGameCheckWin(userState.JoinedSessionId)
	if err != nil {
		return
	}

	if win {
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

func Create(userState *users.UserState) (sessionID string, code string, err error) {
	sessionID = "s-" + uuid.NewString()
	code, err = reserveCode(sessionID)
	if err != nil {
		return
	}
	err = dbSessionCreate(code, sessionID, userState.UserId)
	return
}

func dbSessionCreate(code string, sessionId string, userId string) (err error) {
	logger := logging.Get()
	pipe := msredis.GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

	pipe.Set(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_OWNER_ID), userId, timeToLive)
	pipe.Set(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_GAME_STATE), "LOBBY", timeToLive) // TODO pull LOBBY into constant
	pipe.SetBit(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), 0, 0)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		logger.Error("failed to create a session in db", zap.Error(err), logging.Code(code), logging.SessionId(sessionId), logging.UserId(userId))
		return
	}
	return
}

const MAX_CODE_ATTEMPTS int = 6 // 1-(1000000/(21^6))^6 = 0.999999999, aka almost certain with 1mil codes/day

func dbCodeReserve(sessionId string, code string) (err error) {
	res, err := msredis.GetRedisClient().SetNX(context.TODO(), keys.BuildCodeKey(code), sessionId, time.Hour*24).Result()
	if !res {
		return errors.New("key already exists")
	}
	return
}

func reserveCode(sessionId string) (code string, err error) {
	for i := 0; i < MAX_CODE_ATTEMPTS; i++ {
		code = codes.EncodeRaw(codes.GenerateRandomRaw())
		err = dbCodeReserve(sessionId, code)
		if err == nil {
			return
		}
	}
	panic("Ran out of tries")
}

func DbGameSendVote(userId string, sessionId string, index int32, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	return msredis.GetRedisClient().SetBit(context.TODO(), keys.BuildVotesKey(sessionId, userId), int64(index), voteBit).Err()
}

func DbGameNextVoteInd(sessionId string, userId string) (index int, err error) {

	current := msredis.GetRedisClient().HGet(
		context.TODO(),
		keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND),
		userId,
	)

	if err = current.Err(); err != nil {
		return
	}

	index, err = strconv.Atoi(current.Val())
	if err != nil {
		return
	}

	go func() {
		res := msredis.GetRedisClient().HSet(
			context.TODO(),
			keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND),
			userId,
			index+1,
		)
		if res.Err() != nil {
			logging.Get().Error("failed to increment users vote ind", zap.Error(res.Err()), logging.SessionId(sessionId), logging.UserId(userId))
		}
	}()

	return
}

func GetNextLocForUser(userState *users.UserState) (loc *mealswipepb.Location, err error) {
	ind, err := DbGameNextVoteInd(userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, err = locations.FromInd(userState.JoinedSessionId, int32(ind))
	return
}

func SendNextLocToUser(userState *users.UserState) (err error) {
	loc, err := GetNextLocForUser(userState)
	if err != nil {
		return
	}

	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		Location: loc,
	})
	return
}
