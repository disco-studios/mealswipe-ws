package validators

import (
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Join = []int16{constants.HostState_UNIDENTIFIED}

func ValidateMessageJoin(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Join)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
