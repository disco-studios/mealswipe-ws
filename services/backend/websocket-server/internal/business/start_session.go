package business

import (
	"context"
	"errors"
	"time"
)

func DbStartSession(code string, sessionId string, lat float64, lng float64) (err error) {
	pipe := redisClient.Pipeline()

	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	venueIds, _, err := GrabLocations(lat, lng)
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
