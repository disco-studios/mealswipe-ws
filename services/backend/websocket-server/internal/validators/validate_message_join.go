package validators

import (
	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Join = []int16{core.HostState_UNIDENTIFIED}

func ValidateMessageJoin(userState *core.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Join)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
