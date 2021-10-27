package sessions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/msredis"
)

func getIdFromCode(code string) (sessionId string, err error) {
	key := keys.BuildCodeKey(code)
	result := msredis.GetRedisClient().Get(context.TODO(), key)
	return result.Val(), result.Err()
}

func getActiveUsers(sessionId string) (activeUsers []string, err error) {
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

func getActiveNicknames(sessionId string) (activeNicknames []string, err error) {
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

func joinById(userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
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

func reverse(venues []string) []string {
	for i, j := 0, len(venues)-1; i < j; i, j = i+1, j-1 {
		venues[i], venues[j] = venues[j], venues[i]
	}
	return venues
}

func start(code string, sessionId string, venueIds []string, distances []float64) (err error) {
	pipe := msredis.GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

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

func vote(userId string, sessionId string, index int32, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	return msredis.GetRedisClient().SetBit(context.TODO(), keys.BuildVotesKey(sessionId, userId), int64(index), voteBit).Err()
}

func getWinIndex(sessionId string) (win bool, winningIndex int32, err error) {
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

func create(code string, sessionId string, userId string) (err error) {
	pipe := msredis.GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

	pipe.Set(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_OWNER_ID), userId, timeToLive)
	pipe.Set(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_GAME_STATE), "LOBBY", timeToLive) // TODO pull LOBBY into constant
	pipe.SetBit(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), 0, 0)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}
	return
}

func nextVoteInd(sessionId string, userId string) (index int, err error) {

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
		err = res.Err()
		if err != nil {
			return
		}
	}()

	return
}
