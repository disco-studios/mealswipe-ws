package create

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Create = []int16{constants.HostState_UNIDENTIFIED}

func ValidateMessageCreate(userState *users.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := common.ValidateHostState(userState, AcceptibleHostStates_Create)
	if validateHostError != nil {
		return validateHostError
	}

	nicknameValid, err := common.IsNicknameValid(createMessage.Nickname)
	if err != nil {
		return err
	} else if !nicknameValid {
		return &errors.MessageValidationError{
			MessageType:   "create",
			Clarification: "invalid nickname",
		}
	}

	return
}
