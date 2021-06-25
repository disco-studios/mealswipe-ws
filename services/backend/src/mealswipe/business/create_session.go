package business

import (
	"context"
	"time"
)

func DbCreateSession(code string, sessionId string, userId string) (err error) {
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
