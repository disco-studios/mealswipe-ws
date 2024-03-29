package sessions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.elastic.co/apm"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/msredis"
)

func getIdFromCode(ctx context.Context, code string) (sessionId string, err error) {
	span, ctx := apm.StartSpan(ctx, "getIdFromCode", "sessions")
	defer span.End()

	key := keys.BuildCodeKey(code)
	result := msredis.Get(ctx, key)
	return result.Val(), result.Err()
}

func getActiveUsers(ctx context.Context, sessionId string) (activeUsers []string, err error) {
	span, ctx := apm.StartSpan(ctx, "getActiveUsers", "sessions")
	defer span.End()

	hGetAll := msredis.HGetAll(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_ACTIVE))
	if err = hGetAll.Err(); err != nil {
		err = fmt.Errorf("redis hgetall: %v", err)
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

func getActiveNicknames(ctx context.Context, sessionId string) (activeNicknames []string, err error) {
	span, ctx := apm.StartSpan(ctx, "getActiveNicknames", "sessions")
	defer span.End()

	activeUsers, err := GetActiveUsers(ctx, sessionId)
	if err != nil {
		err = fmt.Errorf("get active users: %v", err)
		return
	}

	hGetAll := msredis.HGetAll(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_NICKNAMES))
	if err = hGetAll.Err(); err != nil {
		err = fmt.Errorf("redis hgetall: %v", err)
		return
	}

	nicknamesMap := hGetAll.Val()
	if err != nil {
		err = fmt.Errorf("nicknamesMap read: %v", err)
		return
	}

	for _, userId := range activeUsers {
		activeNicknames = append(activeNicknames, nicknamesMap[userId])
	}

	return
}

func joinById(ctx context.Context, userId string, sessionId string, nickname string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	span, ctx := apm.StartSpan(ctx, "joinById", "sessions")
	defer span.End()

	pipe := msredis.Pipeline()
	timeToLive := time.Hour * 24

	pipe.SetNX(ctx, keys.BuildUserKey(userId, keys.KEY_USER_SESSION), sessionId, time.Hour*24)
	pipe.SAdd(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS), userId)
	pipe.HSet(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_ACTIVE), userId, true)        // TODO Expire
	pipe.HSet(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND), userId, 0)                // TODO Expire
	pipe.HSet(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS_NICKNAMES), userId, nickname) // TODO Expire
	pipe.SetBit(ctx, keys.BuildVotesKey(sessionId, userId), 0, 0)
	pipe.Expire(ctx, keys.BuildVotesKey(sessionId, userId), timeToLive)

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = fmt.Errorf("redis pipe exec: %v", err)
		return
	}

	// Initiate a pubsub with this session
	redisPubsub = msredis.Subscribe(ctx, keys.BuildSessionKey(sessionId, ""))
	pubsubChannel := redisPubsub.Channel()
	go HandleRedisMessages(pubsubChannel, genericPubsub)

	return
}

func rejoin(ctx context.Context, userId string, sessionId string, genericPubsub chan<- string) (redisPubsub *redis.PubSub, err error) {
	span, ctx := apm.StartSpan(ctx, "rejoin", "sessions")
	defer span.End()

	// Initiate a pubsub with this session
	redisPubsub = msredis.Subscribe(ctx, keys.BuildSessionKey(sessionId, ""))
	pubsubChannel := redisPubsub.Channel()
	go HandleRedisMessages(pubsubChannel, genericPubsub)

	return
}

func isUserInId(ctx context.Context, userId string, sessionId string) (inGame bool, isOwner bool, err error) {
	span, ctx := apm.StartSpan(ctx, "isUserInId", "sessions")
	defer span.End()

	pipe := msredis.Pipeline()
	owner := msredis.Get(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_OWNER_ID))
	isMember := msredis.SIsMember(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_USERS), userId)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return
	}

	return isMember.Val(), owner.Val() == userId, nil
}

func reverse(venues []string) []string {
	for i, j := 0, len(venues)-1; i < j; i, j = i+1, j-1 {
		venues[i], venues[j] = venues[j], venues[i]
	}
	return venues
}

func start(ctx context.Context, code string, sessionId string, venueIds []string, distances []float64) (err error) {
	span, ctx := apm.StartSpan(ctx, "start", "sessions")
	defer span.End()

	pipe := msredis.Pipeline()

	timeToLive := time.Hour * 24

	pipe.Del(ctx, keys.BuildCodeKey(code))
	pipe.Set(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_GAME_STATE), "RUNNING", timeToLive) // TODO pull RUNNING into constant
	pipe.LPush(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), reverse(venueIds))

	var distanceStrings []string
	for _, distance := range distances {
		distanceStrings = append(distanceStrings, fmt.Sprintf("%d", int(distance)))
	}

	pipe.LPush(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), reverse(distanceStrings))
	pipe.Expire(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), timeToLive)
	pipe.Expire(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), timeToLive)

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = fmt.Errorf("redis pipe exec: %v", err)
		return
	}

	return
}

func vote(ctx context.Context, userId string, sessionId string, index int32, state bool) (err error) {
	span, ctx := apm.StartSpan(ctx, "vote", "sessions")
	defer span.End()

	voteBit := 0
	if state {
		voteBit = 1
	}

	return msredis.SetBit(ctx, keys.BuildVotesKey(sessionId, userId), int64(index), voteBit).Err()
}

func getWinIndex(ctx context.Context, sessionId string) (win bool, winningIndex int32, err error) {
	span, ctx := apm.StartSpan(ctx, "getWinIndex", "sessions")
	defer span.End()

	activeUsers, err := GetActiveUsers(ctx, sessionId)
	if err != nil {
		err = fmt.Errorf("get active users: %w", err)
		return
	}

	var voteKeys []string
	for _, userId := range activeUsers {
		voteKeys = append(voteKeys, keys.BuildVotesKey(sessionId, userId))
	}

	pipe := msredis.Pipeline()

	pipe.BitOpAnd(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), voteKeys...)
	winningIndexResult := pipe.BitPos(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), 1)

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = fmt.Errorf("redis pipe exec: %v", err)
		return
	}

	winningIndex = int32(winningIndexResult.Val())
	win = winningIndex > -1
	return
}

func create(ctx context.Context, code string, sessionId string, userId string) (err error) {
	span, ctx := apm.StartSpan(ctx, "create", "sessions")
	defer span.End()

	pipe := msredis.Pipeline()

	timeToLive := time.Hour * 24

	pipe.Set(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_OWNER_ID), userId, timeToLive)
	pipe.Set(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_GAME_STATE), "LOBBY", timeToLive) // TODO pull LOBBY into constant
	pipe.SetBit(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTE_TALLY), 0, 0)

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = fmt.Errorf("redis pipe exec: %v", err)
		return
	}
	return
}

func nextVoteInd(ctx context.Context, sessionId string, userId string) (index int, err error) {
	span, ctx := apm.StartSpan(ctx, "nextVoteInd", "sessions")
	defer span.End()

	// TODO Should we really store this in a set? Probably not
	current := msredis.HGet(
		ctx,
		keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND),
		userId,
	)

	if err = current.Err(); err != nil {
		err = fmt.Errorf("redis hget: %v", err)
		return
	}

	index, err = strconv.Atoi(current.Val())
	if err != nil {
		err = fmt.Errorf("convert string to int: %v", err)
		return
	}

	go func() {
		res := msredis.HSet(
			ctx,
			keys.BuildSessionKey(sessionId, keys.KEY_SESSION_VOTEIND),
			userId,
			index+1,
		)
		err = res.Err()
		if err != nil {
			err = fmt.Errorf("setting redis vote ind: %v", err)
			return
		}
	}()

	return
}
