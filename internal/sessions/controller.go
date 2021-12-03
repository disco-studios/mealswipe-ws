package sessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/codes"
	"mealswipe.app/mealswipe/internal/locations"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/types"
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

	logging.MetricCtx(ctx, "session_join").Info(
		"user joined session",
		zap.String("nickname", userState.Nickname),
	)
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

func Rejoin(ctx context.Context, userState *types.UserState) (inGame bool, isOwner bool, err error) {
	inGame, isOwner, err = isUserInId(ctx, userState.UserId, userState.JoinedSessionId)
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

	activeUsersLen := -1
	activeUsers, err := getActiveUsers(ctx, sessionId)
	if err == nil {
		activeUsersLen = len(activeUsers)
	} else {
		logging.Ctx(ctx).Error("failed to get active users for metric", zap.Error(err))
	}

	logging.MetricCtx(ctx, "session_start").Info(fmt.Sprintf("session %s started", sessionId),
		logging.Code(code),
		zap.Int32("radius", radius),
		zap.String("category", categoryId),
		zap.Int("active_users", activeUsersLen),
		zap.Namespace("geo.location"),
		zap.Float64("lat", lat),
		zap.Float64("lon", lng),
	)
	logging.Get().With()

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
		loc, locId, _err := locations.FromInd(ctx, userState.JoinedSessionId, winIndex)
		if _err != nil {
			err = fmt.Errorf("winning location from ind: %w", _err)
			return
		}

		logging.MetricCtx(ctx, "session_win").Info(fmt.Sprintf("session %s won", userState.JoinedSessionId),
			zap.String("winning_loc", locId),
		)

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

func Create(ctx context.Context, userState *types.UserState) (sessionId string, code string, err error) {
	span, ctx := apm.StartSpan(ctx, "Create", "sessions")
	defer span.End()

	sessionId = "s-" + uuid.NewString()

	logging.MetricCtx(ctx, "session_create").Info(fmt.Sprintf("session %s created", sessionId),
		logging.SessionId(sessionId),
	)

	code, err = codes.Reserve(ctx, sessionId)
	if err != nil {
		return
	}
	err = create(ctx, code, sessionId, userState.UserId)
	return
}

func GetNextLocForUser(ctx context.Context, userState *types.UserState) (loc *mealswipepb.Location, err error) {
	span, ctx := apm.StartSpan(ctx, "GetNextLocForUser", "sessions")
	defer span.End()

	ind, err := nextVoteInd(ctx, userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, _, err = locations.FromInd(ctx, userState.JoinedSessionId, int32(ind))
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
