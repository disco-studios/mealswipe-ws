package business

import (
	"context"
	"errors"
)

func DbGetActiveUsers(sessionId string) (activeUsers []string, err error) {
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

func DbGetNicknames(sessionId string) (nicknames map[string]string, err error) {
	hGetAll := redisClient.HGetAll(context.TODO(), "session."+sessionId+".users.nicknames")
	if err = hGetAll.Err(); err != nil {
		return
	}

	nicknames = hGetAll.Val()
	return
}
