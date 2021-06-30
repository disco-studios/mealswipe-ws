package business

import (
	"context"
	"log"
	"strconv"
)

func DbGameCheckWin(sessionId string) (win bool, winningIndex int64, err error) {
	activeUsers, err := DbSessionGetActiveUsers(sessionId)
	if err != nil {
		return
	}

	var voteKeys []string
	for _, userId := range activeUsers {
		voteKeys = append(voteKeys, "user."+userId+".votes")
	}

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

func DbGameSendVote(userId string, index int64, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	return redisClient.SetBit(context.TODO(), "user."+userId+".votes", index, voteBit).Err()
}

func DbGameNextVoteInd(sessionId string, userId string) (index int, err error) {
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
