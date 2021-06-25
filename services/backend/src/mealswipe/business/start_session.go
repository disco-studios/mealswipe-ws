package business

import (
	"context"
	"time"
)

func DbStartSession(code string, sessionId string, lat float32, lng float32) (err error) {
	pipe := redisClient.Pipeline()

	sessionKey := "session." + sessionId
	timeToLive := time.Hour * 24

	venues, err := GrabLocations(lat, lng)
	if err != nil {
		return
	}

	var venueIds []string
	for _, venue := range venues {
		venueIds = append(venueIds, venue.ID)
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
