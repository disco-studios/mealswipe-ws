package business

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func DbSessionCreate(code string, sessionId string, userId string) (err error) {
	pipe := redisClient.Pipeline()

	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	pipe.Set(context.TODO(), sessionKey+".owner_id", userId, timeToLive)
	pipe.Set(context.TODO(), sessionKey+".game_state", "LOBBY", timeToLive) // TODO pull LOBBY into constant
	pipe.SetBit(context.TODO(), sessionKey+".vote_tally", 0, 0)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}
	return
}

func DbSessionJoinById(userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	pipe := redisClient.Pipeline()
	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	pipe.SetNX(context.TODO(), "user."+userId+".session", sessionId, time.Hour*24)
	pipe.SAdd(context.TODO(), sessionKey+".users", userId)
	pipe.HSet(context.TODO(), sessionKey+".users.active", userId, true)        // TODO Expire
	pipe.HSet(context.TODO(), sessionKey+".users.voteind", userId, 0)          // TODO Expire
	pipe.HSet(context.TODO(), sessionKey+".users.nicknames", userId, nickname) // TODO Expire
	pipe.SetBit(context.TODO(), "user."+userId+".votes", 0, 0)
	pipe.Expire(context.TODO(), "user."+userId+".votes", timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	// Initiate a pubsub with this session
	redisPubsub = redisClient.Subscribe(context.TODO(), "session."+sessionId)
	pubsubChannel := redisPubsub.Channel()
	go handleRedisMessages(pubsubChannel, genericPubsub)

	return
}

func DbSessionGetIdFromCode(code string) (sessionId string, err error) {
	key := "code." + code
	result := redisClient.Get(context.TODO(), key)
	return result.Val(), result.Err()
}

// TODO pull out
func handleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
	log.Println("Redis PubSub cleaned up")
}

func DbSessionStart(code string, sessionId string, lat float64, lng float64) (err error) {
	pipe := redisClient.Pipeline()

	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	venueIds, _, err := DbLocationIdsForLocation(lat, lng)
	if err != nil {
		return
	}
	if len(venueIds) == 0 {
		return errors.New("found no venues for loc")
	}

	pipe.Del(context.TODO(), "code."+code)
	pipe.Set(context.TODO(), sessionKey+".game_state", "RUNNING", timeToLive) // TODO pull RUNNING into constant
	pipe.LPush(context.TODO(), sessionKey+".locations", venueIds)
	pipe.Expire(context.TODO(), sessionKey+".locations", timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}
	return
}

func DbSessionGetActiveUsers(sessionId string) (activeUsers []string, err error) {
	hGetAll := redisClient.HGetAll(context.TODO(), "session."+sessionId+".users.active")
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

func DbSessionGetActiveNicknames(sessionId string) (activeNicknames []string, err error) {
	activeUsers, err := DbSessionGetActiveUsers(sessionId)
	if err != nil {
		return
	}

	hGetAll := redisClient.HGetAll(context.TODO(), "session."+sessionId+".users.nicknames")
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
