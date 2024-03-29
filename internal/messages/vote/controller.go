package vote

import (
	"context"
	"fmt"

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
)

var AcceptibleHostStates_Vote = []int16{mealswipe.HostState_HOSTING, mealswipe.HostState_JOINING}

func HandleMessage(ctx context.Context, userState *types.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	go func() {
		err = sessions.SendNextLocToUser(ctx, userState)
		if err != nil {
			err = fmt.Errorf("send next loc: %w", err)
			logging.Ctx(ctx).Error("failed to send next loc to user", zap.Error(err))
		}
	}()

	err = sessions.Vote(ctx, userState.UserId, userState.JoinedSessionId, voteMessage.Index, voteMessage.Vote)
	if err != nil {
		err = fmt.Errorf("vote: %w", err)
		return
	}

	logging.MetricCtx(ctx, "swipe_dir").Info(
		fmt.Sprintf("voted %t for index %d", voteMessage.Vote, voteMessage.Index),
		zap.Bool("right", voteMessage.Vote),
		zap.Int32("index", voteMessage.Index),
	)
	go sessions.CheckWin(ctx, userState) // TODO This could throw an error, figure out how to handle

	return
}

func ValidateMessage(ctx context.Context, userState *types.UserState, voteMessage *mealswipepb.VoteMessage) (err error, ws_err *mealswipepb.ErrorMessage) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Vote)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return
	}
	return
}
