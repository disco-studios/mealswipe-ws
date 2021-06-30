package business

import (
	"context"
	"errors"
	"time"
)

func DbReserveCode(sessionId string, code string) (err error) {
	res, err := redisClient.SetNX(context.TODO(), "code."+code, sessionId, time.Hour*24).Result()
	if !res {
		return errors.New("key already exists")
	}
	return
}
