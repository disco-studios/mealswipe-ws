package join

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessage(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {

	// Get the session ID for the given code
	sessionId, err := sessions.GetIdFromCode(joinMessage.Code)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = sessions.JoinById(userState, sessionId, joinMessage.Code)
	if err != nil {
		return
	}
	userState.HostState = constants.HostState_JOINING

	// Send the lobby info to the user
	inLobbyNicknames, err := sessions.GetActiveNicknames(userState.JoinedSessionId)

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

var AcceptibleHostStates_Join = []int16{constants.HostState_UNIDENTIFIED}

func ValidateMessage(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := common.ValidateHostState(userState, AcceptibleHostStates_Join)
	if validateHostError != nil {
		return validateHostError
	}

	// Validate that code is valid format
	if !common.IsCodeValid(joinMessage.Code) {
		return &errors.MessageValidationError{
			MessageType:   "join",
			Clarification: "invalid code format",
		}
	}

	// Validate nickname
	nicknameValid, err := common.IsNicknameValid(joinMessage.Nickname)
	if err != nil {
		return err
	} else if !nicknameValid {
		return &errors.MessageValidationError{
			MessageType:   "join",
			Clarification: "invalid nickname",
		}
	}

	// Validate that this session actually exists
	sessionId, err := sessions.GetIdFromCode(joinMessage.Code)
	if err != nil || sessionId == "" {
		return err
	}

	return
}
