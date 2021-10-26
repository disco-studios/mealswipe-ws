package join

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/common/errors"
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Join = []int16{constants.HostState_UNIDENTIFIED}

func ValidateMessageJoin(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
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
	sessionId, err := database.GetIdFromCode(joinMessage.Code)
	if err != nil || sessionId == "" {
		return err
	}

	return
}
