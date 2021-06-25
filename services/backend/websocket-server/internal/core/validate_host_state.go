package core

func ValidateHostState(userState *UserState, allowed []int16) (err error) {
	for _, allowedState := range allowed {
		if userState.HostState == allowedState {
			return
		}
	}
	return &InvalidHostStateError{
		Allowed:  allowed,
		Received: userState.HostState,
	}
}
