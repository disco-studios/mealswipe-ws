package business

import (
	"context"
	"log"
)

func DbCheckWin(sessionId string) (win bool, winningIndex int64, err error) {
	activeUsers, err := DbGetActiveUsers(sessionId)
	if err != nil {
		return
	}

	log.Println("Active", activeUsers)
	var voteKeys []string
	for _, userId := range activeUsers {
		voteKeys = append(voteKeys, "user."+userId+".votes")
	}
	log.Println("Vote keys", voteKeys)

	pipe := redisClient.Pipeline()

	pipe.BitOpAnd(context.TODO(), "session."+sessionId+".vote_tally", voteKeys...)
	winningIndexResult := pipe.BitPos(context.TODO(), "session."+sessionId+".vote_tally", 1)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		return
	}

	winningIndex = winningIndexResult.Val()
	win = winningIndex > -1
	return
}
