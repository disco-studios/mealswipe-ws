package validators

import (
	"log"
	"testing"

	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func TestValidateJoinMessage(t *testing.T) {
	userState := users.CreateUserState()
	joinMessage := &mealswipepb.JoinMessage{
		Nickname: "Cam the Man",
		Code:     "ABCDEF",
	}

	t.Run("HostState_UNIDENTIFIED valid", func(t *testing.T) {
		if err := ValidateMessageJoin(userState, joinMessage); err != nil {
			log.Fatal(err)
		}
	})
	t.Run("HostState_JOINING invalid", func(t *testing.T) {
		userState.HostState = constants.HostState_JOINING
		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
			t.FailNow()
		}
	})
	t.Run("HostState_HOSTING invalid", func(t *testing.T) {
		userState.HostState = constants.HostState_HOSTING
		if err := ValidateMessageJoin(userState, joinMessage); err == nil {
			t.FailNow()
		}
	})

}
