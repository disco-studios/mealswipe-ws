package validators

import (
	"log"
	"testing"

	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func TestValidateStartMessage(t *testing.T) {
	userState := core.CreateUserState()
	startMessage := &mealswipepb.StartMessage{
		Lat: 0.0,
		Lng: 0.0,
	}

	t.Run("HostState_UNIDENTIFIED invalid", func(t *testing.T) {
		if err := ValidateMessageStart(userState, startMessage); err == nil {
			t.FailNow()
		}
	})
	t.Run("HostState_JOINING invalid", func(t *testing.T) {
		userState.HostState = core.HostState_JOINING
		if err := ValidateMessageStart(userState, startMessage); err == nil {
			t.FailNow()
		}
	})
	t.Run("HostState_HOSTING valid", func(t *testing.T) {
		userState.HostState = core.HostState_HOSTING
		if err := ValidateMessageStart(userState, startMessage); err != nil {
			log.Fatal(err)
		}
	})

}
