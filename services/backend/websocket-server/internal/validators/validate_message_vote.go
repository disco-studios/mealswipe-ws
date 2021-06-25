package validators

import (
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Vote = []int16{core.HostState_HOSTING, core.HostState_JOINING}

func ValidateMessageVote(userState *core.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := core.ValidateHostState(userState, AcceptibleHostStates_Vote)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
