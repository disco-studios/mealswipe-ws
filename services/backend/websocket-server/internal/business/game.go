package business

import (
	"context"
	"log"
	"strconv"
)

func DbGameCheckWin(sessionId string) (win bool, winningIndex int64, err error) {
	activeUsers, err := DbSessionGetActiveUsers(sessionId)
	if err != nil {
		log.Print("can't get active users")
		return
	}

	var voteKeys []string
	for _, userId := range activeUsers {
		voteKeys = append(voteKeys, BuildVotesKey(sessionId, userId))
	}

	pipe := GetRedisClient().Pipeline()

	pipe.BitOpAnd(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_VOTE_TALLY), voteKeys...)
	winningIndexResult := pipe.BitPos(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_VOTE_TALLY), 1)

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Print("can't tally votes")
		return
	}

	winningIndex = winningIndexResult.Val()
	win = winningIndex > -1
	return
}

func DbGameSendVote(userId string, sessionId string, index int64, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	// Register statistics async
	go StatsRegisterSwipe(sessionId, index, state)

	return GetRedisClient().SetBit(context.TODO(), BuildVotesKey(sessionId, userId), index, voteBit).Err()
}

func DbGameNextVoteInd(sessionId string, userId string) (index int, err error) {
	current := GetRedisClient().HGet(
		context.TODO(),
		BuildSessionKey(sessionId, KEY_SESSION_VOTEIND),
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
		res := GetRedisClient().HSet(
			context.TODO(),
			BuildSessionKey(sessionId, KEY_SESSION_VOTEIND),
			userId,
			index+1,
		)
		if res.Err() != nil {
			log.Println("failed to increment:", res.Err())
		}
	}()

	return
}
