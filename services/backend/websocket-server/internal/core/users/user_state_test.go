package users

import (
	"testing"

	"mealswipe.app/mealswipe/internal/common/constants"
)

func TestCreateUserState(t *testing.T) {
	userState := CreateUserState()

	t.Run("user state starts with unidentified host state", func(t *testing.T) {
		if userState.HostState != constants.HostState_UNIDENTIFIED {
			t.FailNow()
		}
	})
}
