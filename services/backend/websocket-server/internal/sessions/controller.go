package sessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/codes"
	"mealswipe.app/mealswipe/internal/locations"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GetIdFromCode(ctx context.Context, code string) (sessionId string, err error) {
	span, ctx := apm.StartSpan(ctx, "GetIdFromCode", "sessions")
	defer span.End()

	return getIdFromCode(ctx, code)
}

func GetActiveUsers(ctx context.Context, sessionId string) (activeUsers []string, err error) {
	span, ctx := apm.StartSpan(ctx, "GetActiveUsers", "sessions")
	defer span.End()

	return getActiveUsers(ctx, sessionId)
}

func GetActiveNicknames(ctx context.Context, sessionId string) (activeNicknames []string, err error) {
	span, ctx := apm.StartSpan(ctx, "GetActiveNicknames", "sessions")
	defer span.End()

	return getActiveNicknames(ctx, sessionId)
}

func JoinById(ctx context.Context, userState *types.UserState, sessionId string, code string) (err error) {
	span, ctx := apm.StartSpan(ctx, "JoinById", "sessions")
	defer span.End()

	logging.ApmCtx(ctx).Info("user joined session", logging.Metric("session_join"), zap.String("nickname", userState.Nickname), logging.UserId(userState.UserId), logging.SessionId(sessionId))
	redisPubsub, err := joinById(ctx, userState.UserId, sessionId, userState.Nickname, userState.PubsubChannel)
	if err != nil {
		err = fmt.Errorf("join by id: %w", err)
		return
	}

	userState.RedisPubsub = redisPubsub
	userState.JoinedSessionId = sessionId
	userState.JoinedSessionCode = code

	return
}

func Rejoin(ctx context.Context, userState *types.UserState) (inGame bool, err error) {
	inGame, err = isUserInId(ctx, userState.UserId, userState.JoinedSessionId)
	if (err != nil) || (!inGame) {
		return
	}

	redisPubsub, err := rejoin(ctx, userState.UserId, userState.JoinedSessionId, userState.PubsubChannel)
	if err != nil {
		err = fmt.Errorf("rejoin by id: %w", err)
		return
	}
	userState.RedisPubsub = redisPubsub

	return
}

// TODO This shouldn't live here
func HandleRedisMessages(redisPubsub <-chan *redis.Message, genericPubsub chan<- string) {
	for msg := range redisPubsub {
		genericPubsub <- msg.Payload
	}
}

func Start(ctx context.Context, code string, sessionId string, lat float64, lng float64, radius int32, categoryId string) (err error) {
	span, ctx := apm.StartSpan(ctx, "Start", "sessions")
	defer span.End()

	logging.ApmCtx(ctx).Info(
		"game started",
		logging.Metric("game_start"),
		logging.Code(code),
		logging.SessionId(sessionId),
		zap.Float64("lat", lat),
		zap.Float64("lng", lng),
		zap.Int32("radius", radius),
		zap.String("category", categoryId),
	)

	// TODO This write should maybe go into locs
	venueIds, distances, err := locations.IdsForLocation(ctx, lat, lng, radius, categoryId)
	if err != nil {
		err = fmt.Errorf("get ids for location: %w", err)
		return
	}
	if len(venueIds) == 0 {
		return errors.New("found no venues for loc")
	}

	return start(ctx, code, sessionId, venueIds, distances)
}

func Vote(ctx context.Context, userId string, sessionId string, index int32, state bool) (err error) {
	span, ctx := apm.StartSpan(ctx, "Vote", "sessions")
	defer span.End()

	return vote(ctx, userId, sessionId, index, state)
}

func CheckWin(ctx context.Context, userState *types.UserState) (err error) {
	span, ctx := apm.StartSpan(ctx, "CheckWin", "sessions")
	defer span.End()

	win, winIndex, err := getWinIndex(ctx, userState.JoinedSessionId)
	if err != nil {
		err = fmt.Errorf("check win index: %w", err)
		return
	}

	if win {
		// TODO This shouldn't live here I don't think
		var loc *mealswipepb.Location
		loc, err = locations.FromInd(ctx, userState.JoinedSessionId, winIndex)
		if err != nil {
			err = fmt.Errorf("winning location from ind: %w", err)
			return
		}
		// TODO Log winning loc

		err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
			GameWinMessage: &mealswipepb.GameWinMessage{
				Locations: []*mealswipepb.WinningLocation{
					{
						Location: loc,
						Votes:    0, // TODO: Impl
					},
				},
			},
		})
		if err != nil {
			err = fmt.Errorf("sending winning message: %w", err)
			return
		}
	}
	return
}

func Create(ctx context.Context, userState *types.UserState) (sessionID string, code string, err error) {
	span, ctx := apm.StartSpan(ctx, "Create", "sessions")
	defer span.End()

	sessionID = "s-" + uuid.NewString()
	code, err = codes.Reserve(ctx, sessionID)
	if err != nil {
		return
	}
	err = create(ctx, code, sessionID, userState.UserId)
	return
}

func GetNextLocForUser(ctx context.Context, userState *types.UserState) (loc *mealswipepb.Location, err error) {
	span, ctx := apm.StartSpan(ctx, "GetNextLocForUser", "sessions")
	defer span.End()

	ind, err := nextVoteInd(ctx, userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, err = locations.FromInd(ctx, userState.JoinedSessionId, int32(ind))
	return
}

func SendNextLocToUser(ctx context.Context, userState *types.UserState) (err error) {
	span, ctx := apm.StartSpan(ctx, "SendNextLocToUser", "sessions")
	defer span.End()

	loc, err := GetNextLocForUser(ctx, userState)
	if err != nil {
		return
	}

	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		Location: loc,
	})
	return
}
