package vote

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Vote = []int16{mealswipe.HostState_HOSTING, mealswipe.HostState_JOINING}

func HandleMessage(ctx context.Context, userState *types.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	logger := logging.Get()
	err = sessions.Vote(ctx, userState.UserId, userState.JoinedSessionId, voteMessage.Index, voteMessage.Vote)
	if err != nil {
		err = fmt.Errorf("vote: %w", err)
		return
	}

	logger.Info("user_vote",
		logging.Metric("swipe_dir"),
		zap.Bool("right", voteMessage.Vote),
		logging.SessionId(userState.JoinedSessionId),
		zap.Int32("index", voteMessage.Index),
		logging.UserId(userState.UserId),
	)
	go sessions.CheckWin(ctx, userState) // TODO This could throw an error, figure out how to handle

	err = sessions.SendNextLocToUser(ctx, userState)
	if err != nil {
		err = fmt.Errorf("send next loc: %w", err)
		return
	}

	return
}

func ValidateMessage(userState *types.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Vote)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return
	}
	return
}
