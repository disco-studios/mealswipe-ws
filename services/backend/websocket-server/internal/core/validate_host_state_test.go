package core

import (
	"log"
	"testing"
)

// TODO Make sure errors are of right type
func TestValidateCreateMessage(t *testing.T) {
	userState := CreateUserState()

	t.Run("valid in list", func(t *testing.T) {
		userState.HostState = HostState_HOSTING
		if err := ValidateHostState(userState, []int16{
			HostState_HOSTING,
			HostState_JOINING,
		}); err != nil {
			log.Fatal(err)
		}
	})

	t.Run("valid single", func(t *testing.T) {
		userState.HostState = HostState_UNIDENTIFIED
		if err := ValidateHostState(userState, []int16{
			HostState_UNIDENTIFIED,
		}); err != nil {
			log.Fatal(err)
		}
	})

	t.Run("invalid in list", func(t *testing.T) {
		userState.HostState = HostState_UNIDENTIFIED
		if err := ValidateHostState(userState, []int16{
			HostState_HOSTING,
			HostState_JOINING,
		}); err == nil {
			t.FailNow()
		}
	})

	t.Run("invalid single", func(t *testing.T) {
		userState.HostState = HostState_HOSTING
		if err := ValidateHostState(userState, []int16{
			HostState_UNIDENTIFIED,
		}); err == nil {
			t.FailNow()
		}
	})
}
