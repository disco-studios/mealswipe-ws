package vote

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

var AcceptibleHostStates_Vote = []int16{constants.HostState_HOSTING, constants.HostState_JOINING}

func ValidateMessageVote(userState *users.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	// Validate that the user is in a state that can do this action
	validateHostError := common.ValidateHostState(userState, AcceptibleHostStates_Vote)
	if validateHostError != nil {
		return validateHostError
	}
	return
}
