package join

import (
	"fmt"

	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessage(userState *types.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {

	// Get the session ID for the given code
	sessionId, err := sessions.GetIdFromCode(joinMessage.Code)
	if err != nil {
		err = fmt.Errorf("get id from code: %w", err)
		return
	}

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = sessions.JoinById(userState, sessionId, joinMessage.Code)
	if err != nil {
		err = fmt.Errorf("join by id: %w", err)
		return
	}
	userState.HostState = mealswipe.HostState_JOINING

	// Send the lobby info to the user
	inLobbyNicknames, err := sessions.GetActiveNicknames(userState.JoinedSessionId)
	if err != nil {
		err = fmt.Errorf("get active nicknames for lobby info: %w", err)
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

func ValidateMessage(userState *types.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
	// Validate that the user is in a state that can do this action
	err = common.ValidateHostState(userState, AcceptibleHostStates_Join)
	if err != nil {
		err = fmt.Errorf("validate host state: %w", err)
		return err
	}

	// Validate that code is valid format
	if !common.IsCodeValid(joinMessage.Code) {
		return &mealswipe.MessageValidationError{
			MessageType:   "join",
			Clarification: "invalid code format",
		}
	}

	// Validate nickname
	nicknameValid, err := common.IsNicknameValid(joinMessage.Nickname)
	if err != nil {
		err = fmt.Errorf("validate nickname: %w", err)
		return err
	} else if !nicknameValid {
		return &mealswipe.MessageValidationError{
			MessageType:   "join",
			Clarification: "invalid nickname",
		}
	}

	// Validate that this session actually exists
	sessionId, err := sessions.GetIdFromCode(joinMessage.Code)
	if err != nil || sessionId == "" {
		err = fmt.Errorf("get session id from code: %w", err)
		return err
	}

	return
}
