package create

import (
	"fmt"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Create = []int16{mealswipe.HostState_UNIDENTIFIED}

func HandleMessage(userState *types.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Create session
	sessionId, code, err := sessions.Create(userState)
	if err != nil {
		err = fmt.Errorf("create session: %w", err)
		return
	}

	// Join the user into the new session
	userState.Nickname = createMessage.Nickname
	err = sessions.JoinById(userState, sessionId, code)
	if err != nil {
		err = fmt.Errorf("join session by id: %w", err)
		return
	}

	userState.HostState = mealswipe.HostState_HOSTING

	// Send the lobby info to the user
	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:     code,
			Nickname: userState.Nickname,
			Users:    []string{userState.Nickname},
		},
	})
	return
}

func ValidateMessage(userState *types.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Create)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err
	}

	nicknameValid, err := common.IsNicknameValid(createMessage.Nickname)
	if err != nil {
		err = fmt.Errorf("validate nickname: %w", err)
		return err
	} else if !nicknameValid {
		logging.Get().Info("invalid nickname given", logging.Metric("bad_nickname"), zap.String("nickname", createMessage.Nickname))
		return &mealswipe.MessageValidationError{
			MessageType:   "create",
			Clarification: "invalid nickname",
		}
	}

	return
}
