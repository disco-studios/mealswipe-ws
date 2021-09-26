package business

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func DbSessionCreate(code string, sessionId string, userId string) (err error) {
	pipe := GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

	pipe.Set(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_OWNER_ID), userId, timeToLive)
	pipe.Set(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_GAME_STATE), "LOBBY", timeToLive) // TODO pull LOBBY into constant
	pipe.SetBit(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_VOTE_TALLY), 0, 0)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't create")
		return
	}
	return
}

func DbSessionJoinById(userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	pipe := GetRedisClient().Pipeline()
	timeToLive := time.Hour * 24

	pipe.SetNX(context.TODO(), BuildUserKey(userId, KEY_USER_SESSION), sessionId, time.Hour*24)
	pipe.SAdd(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_USERS), userId)
	pipe.HSet(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_USERS_ACTIVE), userId, true)        // TODO Expire
	pipe.HSet(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_VOTEIND), userId, 0)                // TODO Expire
	pipe.HSet(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_USERS_NICKNAMES), userId, nickname) // TODO Expire
	pipe.SetBit(context.TODO(), BuildVotesKey(sessionId, userId), 0, 0)
	pipe.Expire(context.TODO(), BuildVotesKey(sessionId, userId), timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't join by id")
		return
	}

	// Initiate a pubsub with this session
	redisPubsub = GetRedisClient().Subscribe(context.TODO(), BuildSessionKey(sessionId, ""))
	pubsubChannel := redisPubsub.Channel()
	go handleRedisMessages(pubsubChannel, genericPubsub)

	return
}

func DbSessionGetIdFromCode(code string) (sessionId string, err error) {
	key := BuildCodeKey(code)
	result := GetRedisClient().Get(context.TODO(), key)
	return result.Val(), result.Err()
}

// TODO pull out
func handleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
	log.Println("Redis PubSub cleaned up")
}

func reverseVenueIds(venues []string) []string {
	for i, j := 0, len(venues)-1; i < j; i, j = i+1, j-1 {
		venues[i], venues[j] = venues[j], venues[i]
	}
	return venues
}

func DbSessionStart(code string, sessionId string, lat float64, lng float64, radius int32) (err error) {
	pipe := GetRedisClient().Pipeline()

	timeToLive := time.Hour * 24

	venueIds, _, err := DbLocationIdsForLocation(lat, lng, radius)
	if err != nil {
		log.Println("can't get start locations")
		return
	}
	if len(venueIds) == 0 {
		return errors.New("found no venues for loc")
	}

	pipe.Del(context.TODO(), BuildCodeKey(code))
	pipe.Set(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_GAME_STATE), "RUNNING", timeToLive) // TODO pull RUNNING into constant
	pipe.LPush(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), reverseVenueIds(venueIds))
	pipe.Expire(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), timeToLive)

	pipout, err := pipe.Exec(context.TODO())
	if err != nil {
		log.Println(BuildCodeKey(code), pipout[0].Err())
		log.Println(BuildSessionKey(sessionId, KEY_SESSION_GAME_STATE), pipout[1].Err())
		log.Println(BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), pipout[2].Err())
		log.Println(BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), pipout[3].Err())
		log.Println("can't start")
		return
	}

	// Register statistics async
	go StatsRegisterGameStart(sessionId)

	return
}

func DbSessionGetActiveUsers(sessionId string) (activeUsers []string, err error) {
	hGetAll := GetRedisClient().HGetAll(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_USERS_ACTIVE))
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
		log.Println("can't get active users")
		return
	}

	hGetAll := GetRedisClient().HGetAll(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_USERS_NICKNAMES))
	if err = hGetAll.Err(); err != nil {
		log.Println("can't get nicknames")
		return
	}

	nicknamesMap := hGetAll.Val()
	if err != nil {
		log.Println("can't get nicknames val")
		return
	}

	for _, userId := range activeUsers {
		activeNicknames = append(activeNicknames, nicknamesMap[userId])
	}
	return
}
