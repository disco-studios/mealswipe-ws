package business

import (
	"context"
	"log"
	"strconv"
)

func DbGrabAndIncrVoteInd(sessionId string, userId string) (index int, err error) {
	current := redisClient.HGet(
		context.TODO(),
		"session."+sessionId+".users.voteind",
		userId,
	)

	if err = current.Err(); err != nil {
		log.Print("can't get ind")
		return
	}

	index, err = strconv.Atoi(current.Val())
	if err != nil {
		return
	}

	go func() {
		res := redisClient.HSet(
			context.TODO(),
			"session."+sessionId+".users.voteind",
			userId,
			index+1,
		)
		if res.Err() != nil {
			log.Println("failed to increment:", res.Err())
		}
	}()

	return
}
