package validators

import (
	"log"
	"testing"

	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

func TestValidateCreateMessage(t *testing.T) {
	userState := core.CreateUserState()
	createMessage := &mealswipepb.CreateMessage{
		Nickname: "Cam the Man",
	}

	t.Run("HostState_UNIDENTIFIED valid", func(t *testing.T) {
		if err := ValidateMessageCreate(userState, createMessage); err != nil {
			log.Fatal(err)
		}
	})
	t.Run("HostState_JOINING invalid", func(t *testing.T) {
		userState.HostState = core.HostState_JOINING
		if err := ValidateMessageCreate(userState, createMessage); err == nil {
			t.FailNow()
		}
	})
	t.Run("HostState_HOSTING invalid", func(t *testing.T) {
		userState.HostState = core.HostState_HOSTING
		if err := ValidateMessageCreate(userState, createMessage); err == nil {
			t.FailNow()
		}
	})

}
