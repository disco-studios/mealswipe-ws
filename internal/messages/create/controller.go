package create

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

var AcceptibleHostStates_Create = []int16{mealswipe.HostState_UNIDENTIFIED}

func HandleMessage(ctx context.Context, userState *types.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Create session
	sessionId, code, err := sessions.Create(ctx, userState)
	if err != nil {
		err = fmt.Errorf("create session: %w", err)
		return
	}

	ctx = userState.TagContext(ctx)

	// Join the user into the new session
	userState.Nickname = createMessage.Nickname
	err = sessions.JoinById(ctx, userState, sessionId, code)
	if err != nil {
		err = fmt.Errorf("join session by id: %w", err)
		return
	}

	userState.HostState = mealswipe.HostState_HOSTING

	// Send the lobby info to the user
	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:      code,
			Nickname:  userState.Nickname,
			Users:     []string{userState.Nickname},
			SessionId: userState.JoinedSessionId,
			UserId:    userState.UserId,
		},
	})
	return
}

func ValidateMessage(ctx context.Context, userState *types.UserState, createMessage *mealswipepb.CreateMessage) (err error, ws_error *mealswipepb.ErrorMessage) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Create)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err, nil
	}

	nicknameValid, err := common.IsNicknameValid(createMessage.Nickname)
	if err != nil {
		err = fmt.Errorf("validate nickname: %w", err)
		return err, nil
	} else if !nicknameValid {
		logging.MetricCtx(ctx, "bad_nickname").Info(
			fmt.Sprintf("gave bad nickname %s", createMessage.Nickname),
			zap.String("nickname", createMessage.Nickname),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "create",
				Clarification: "invalid nickname",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidNicknameError,
				Message:   "TODO: More info", // TODO More info
			}
	}

	return
}
