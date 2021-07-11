package business

import (
	"context"
	"errors"
	"time"
)

func DbCodeReserve(sessionId string, code string) (err error) {
	res, err := GetRedisClient().SetNX(context.TODO(), "code."+code, sessionId, time.Hour*24).Result()
	if !res {
		return errors.New("key already exists")
	}
	return
}
