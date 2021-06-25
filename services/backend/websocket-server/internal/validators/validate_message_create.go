package validators

import (
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Create = []int16{core.HostState_UNIDENTIFIED}

func ValidateMessageCreate(userState *core.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Create)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
