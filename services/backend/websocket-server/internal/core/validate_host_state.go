package core

import (
	"mealswipe.app/mealswipe/internal/common/errors"
	"mealswipe.app/mealswipe/internal/core/users"
)

func ValidateHostState(userState *users.UserState, allowed []int16) (err error) {
	for _, allowedState := range allowed {
		if userState.HostState == allowedState {
			return
		}
	}
	return &errors.InvalidHostStateError{
		Allowed:  allowed,
		Received: userState.HostState,
	}
}
