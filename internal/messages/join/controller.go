package join

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

func HandleMessage(ctx context.Context, userState *types.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
	userState.UserId = joinMessage.Uuid
	ctx = userState.TagContext(ctx)

	// Get the session ID for the given code
	sessionId, _err := sessions.GetIdFromCode(ctx, joinMessage.Code)
	if _err != nil {
		err = fmt.Errorf("get id from code: %w", _err)
		return
	}

	ctx = userState.TagContext(ctx)

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = sessions.JoinById(ctx, userState, sessionId, joinMessage.Code)
	if err != nil {
		err = fmt.Errorf("join by id: %w", err)
		return
	}
	userState.HostState = mealswipe.HostState_JOINING

	// Send the lobby info to the user
	inLobbyNicknames, _err := sessions.GetActiveNicknames(ctx, userState.JoinedSessionId)
	if _err != nil {
		err = fmt.Errorf("get active nicknames for lobby info: %w", _err)
	}

	// Broadcast user join
	userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:     joinMessage.Code,
			Nickname: userState.Nickname,
			Users:    inLobbyNicknames,
		},
	})

	return
}

var AcceptibleHostStates_Join = []int16{mealswipe.HostState_UNIDENTIFIED}

func ValidateMessage(ctx context.Context, userState *types.UserState, joinMessage *mealswipepb.JoinMessage) (err error, ws_error *mealswipepb.ErrorMessage) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Join)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err, nil
	}

	// Validate that code is valid format
	if !common.IsCodeValid(joinMessage.Code) {
		logging.MetricCtx(ctx, "bad_code").Info(
			fmt.Sprintf("invalid code given for %s", userState.UserId),
			zap.String("code", joinMessage.Code),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "join",
				Clarification: "invalid code format",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidCodeError,
				Message:   fmt.Sprintf("Could not find a code %s", joinMessage.Code),
			}
	}

	// Validate nickname
	nicknameValid, err := common.IsNicknameValid(joinMessage.Nickname)
	if err != nil {
		err = fmt.Errorf("validate nickname: %w", err)
		return err, nil
	} else if !nicknameValid {
		logging.MetricCtx(ctx, "bad_nickname").Info(
			"invalid nickname given",
			zap.String("nickname", joinMessage.Nickname),
		)
		return &mealswipe.MessageValidationError{
				MessageType:   "join",
				Clarification: "invalid nickname",
			}, &mealswipepb.ErrorMessage{
				ErrorType: mealswipepb.ErrorType_InvalidNicknameError,
				Message:   "TODO: More info", // TODO More info
			}
	}

	if joinMessage.Uuid != "" {
		// if !uuidRegex.Match([]byte(joinMessage.Uuid)) {
		// 	logging.MetricCtx(ctx, "bad_uuid").Info(
		// 		"invalid uuid given",
		// 		zap.String("uuid", joinMessage.Uuid),
		// 	)
		// 	return &mealswipe.MessageValidationError{
		// 		MessageType:   "join",
		// 		Clarification: "invalid uuid",
		// 	}
		// }
	}

	// Validate that this session actually exists
	sessionId, err := sessions.GetIdFromCode(ctx, joinMessage.Code)
	if err != nil || sessionId == "" {
		err = fmt.Errorf("get session id from code: %w", err)
		return err, nil
	}

	return
}
