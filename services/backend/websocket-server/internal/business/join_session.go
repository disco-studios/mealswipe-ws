package business

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// TODO pull out
func handleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
	log.Println("Redis PubSub cleaned up")
}

func DbUserJoinSessionById(userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	pipe := redisClient.Pipeline()
	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	pipe.SetNX(context.TODO(), "user."+userId+".session", sessionId, time.Hour*24)
	pipe.SAdd(context.TODO(), sessionKey+".users", userId)
	pipe.SetNX(context.TODO(), "user."+userId+".nickname", nickname, timeToLive)
	pipe.HSet(context.TODO(), sessionKey+".users.active", userId, true)
	pipe.HSet(context.TODO(), sessionKey+".users.voteind", userId, 0)
	pipe.SetBit(context.TODO(), "user."+userId+".votes", 0, 0)
	pipe.Expire(context.TODO(), "user."+userId+".votes", timeToLive)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	// Initiate a pubsub with this session
	redisPubsub = redisClient.Subscribe(ctx, "session."+sessionId)
	pubsubChannel := redisPubsub.Channel()
	go handleRedisMessages(pubsubChannel, genericPubsub)

	return
}
