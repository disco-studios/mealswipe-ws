package rejoin

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
)

var uuidRegex, _ = regexp.Compile("^u-[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}$")

func HandleMessage(ctx context.Context, userState *types.UserState, rejoinMessage *mealswipepb.RejoinMessage) (err error) {
	userState.JoinedSessionId = rejoinMessage.SessionId
	userState.UserId = rejoinMessage.UserId

	inGame, isOwner, _err := sessions.Rejoin(ctx, userState)
	if _err != nil {
		err = fmt.Errorf("rejoin: %w", _err)
		return
	}

	if !inGame {
		err = fmt.Errorf("bad rejoin info given")
	}

	if isOwner {
		userState.HostState = mealswipe.HostState_HOSTING
	} else {
		userState.HostState = mealswipe.HostState_JOINING
	}

	return
}

var AcceptibleHostStates_Join = []int16{mealswipe.HostState_UNIDENTIFIED}

func ValidateMessage(ctx context.Context, userState *types.UserState, rejoinMessage *mealswipepb.RejoinMessage) (err error, ws_err *mealswipepb.ErrorMessage) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Join)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err, nil
	}

	if rejoinMessage.UserId == "" {
		logging.MetricCtx(ctx, "bad_user_id").Info(
			"rejoin: invalid userid given",
			logging.UserId(rejoinMessage.UserId),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "rejoin",
				Clarification: "invalid userid format",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidUserIdError,
				Message:   "Must provide a user ID",
			}
	}

	if rejoinMessage.SessionId == "" {
		logging.MetricCtx(ctx, "bad_session_id").Info(
			"rejoin: invalid sessionid given",
			logging.SessionId(rejoinMessage.SessionId),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "rejoin",
				Clarification: "invalid sessionid format",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidSessionIdError,
				Message:   "Invalid session ID format",
			}
	}

	if !uuidRegex.Match([]byte(rejoinMessage.UserId)) {
		logging.MetricCtx(ctx, "bad_uuid_rejoin").Info(
			"invalid uuid given",
			zap.String("uuid", rejoinMessage.UserId),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "rejoin",
				Clarification: "invalid uuid",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidUserIdError,
				Message:   "Invalid user ID format",
			}
	}

	return
}
