package validators

import (
	"regexp"

	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Create = []int16{constants.HostState_UNIDENTIFIED}

func ValidateMessageCreate(userState *users.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Create)
	if validateHostError != nil {
		return validateHostError
	}

	nicknameValid, err := isNicknameValid(createMessage.Nickname)
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

func isNicknameValid(nickname string) (valid bool, err error) {
	if len(nickname) == 0 || len(nickname) > 16 {
		return false, nil
	}
	// - Does not start or end with a space
	// - Only contains a-zA-Z and space
	// - Can only have one space in a row
	return regexp.MatchString(`^([a-zA-Z]+ ?)*[a-zA-Z]$`, nickname)
}
