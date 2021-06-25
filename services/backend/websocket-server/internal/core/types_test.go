package core

import (
	"testing"
)

func TestCreateUserState(t *testing.T) {
	userState := CreateUserState()

	t.Run("user state starts with unidentified host state", func(t *testing.T) {
		if userState.HostState != HostState_UNIDENTIFIED {
			t.FailNow()
		}
	})
}
