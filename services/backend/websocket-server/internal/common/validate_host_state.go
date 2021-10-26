package common

import (
	"mealswipe.app/mealswipe/internal/types"
	"mealswipe.app/mealswipe/pkg/mealswipe"
)

func ValidateHostState(userState *types.UserState, allowed []int16) (err error) {
	for _, allowedState := range allowed {
		if userState.HostState == allowedState {
			return
		}
	}
	return &mealswipe.InvalidHostStateError{
		Allowed:  allowed,
		Received: userState.HostState,
	}
}
