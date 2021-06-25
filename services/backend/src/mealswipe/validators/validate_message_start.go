package validators

import (
	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Start = []int16{core.HostState_HOSTING}

func ValidateMessageStart(userState *core.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Start)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
