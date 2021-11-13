package rejoin

import (
	"context"
	"fmt"

	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessage(ctx context.Context, userState *types.UserState, rejoinMessage *mealswipepb.RejoinMessage) (err error) {
	userState.JoinedSessionId = rejoinMessage.SessionId
	userState.UserId = rejoinMessage.UserId

	inGame, _err := sessions.Rejoin(ctx, userState)
	if _err != nil {
		err = fmt.Errorf("rejoin: %w", _err)
		return
	}

	if !inGame {
		err = fmt.Errorf("bad rejoin info given")
	}

	userState.HostState = mealswipe.HostState_JOINING

	return
}

var AcceptibleHostStates_Join = []int16{mealswipe.HostState_UNIDENTIFIED}

func ValidateMessage(ctx context.Context, userState *types.UserState, rejoinMessage *mealswipepb.RejoinMessage) (err error) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Join)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err
	}

	if rejoinMessage.UserId == "" {
		logging.ApmCtx(ctx).Info("rejoin: invalid userid given", logging.Metric("bad_user_id"), logging.UserId(rejoinMessage.UserId))
		return &mealswipe.MessageValidationError{
			MessageType:   "rejoin",
			Clarification: "invalid userid format",
		}
	}

	if rejoinMessage.SessionId == "" {
		logging.ApmCtx(ctx).Info("rejoin: invalid sessionid given", logging.Metric("bad_session_id"), logging.SessionId(rejoinMessage.SessionId))
		return &mealswipe.MessageValidationError{
			MessageType:   "rejoin",
			Clarification: "invalid sessionid format",
		}
	}

	return
}
