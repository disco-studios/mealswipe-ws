package business

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common/logging"
)

func DbGameCheckWin(sessionId string) (win bool, winningIndex int32, err error) {
	logger := logging.Get()
	activeUsers, err := DbSessionGetActiveUsers(sessionId)
	if err != nil {
		logger.Error("can't get active users", zap.Error(err), logging.SessionId(sessionId))
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
		logger.Error("can't tally votes", zap.Error(err), logging.SessionId(sessionId))
		return
	}

	winningIndex = int32(winningIndexResult.Val())
	win = winningIndex > -1
	return
}

func DbGameSendVote(userId string, sessionId string, index int32, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	// Register statistics async
	go StatsRegisterSwipe(sessionId, index, state)

	return GetRedisClient().SetBit(context.TODO(), BuildVotesKey(sessionId, userId), int64(index), voteBit).Err()
}

func DbGameNextVoteInd(sessionId string, userId string) (index int, err error) {
	logger := logging.Get()

	current := GetRedisClient().HGet(
		context.TODO(),
		BuildSessionKey(sessionId, KEY_SESSION_VOTEIND),
		userId,
	)

	if err = current.Err(); err != nil {
		logger.Error("can't get next vote ind", zap.Error(err), logging.SessionId(sessionId), logging.UserId(userId))
		return
	}

	index, err = strconv.Atoi(current.Val())
	if err != nil {
		logger.Error("failed to parse next vote ind to int", zap.Error(err), logging.SessionId(sessionId), logging.UserId(userId), zap.String("input", current.Val()))
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
			logger.Error("failed to increment users vote ind", zap.Error(res.Err()), logging.SessionId(sessionId), logging.UserId(userId))
		}
	}()

	return
}
