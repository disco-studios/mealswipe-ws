package validators

import (
	"log"
	"testing"

	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func TestValidateVoteMessage(t *testing.T) {
	userState := users.CreateUserState()
	voteMessage := &mealswipepb.VoteMessage{
		Index: 0,
		Vote:  true,
	}

	t.Run("HostState_UNIDENTIFIED invalid", func(t *testing.T) {
		if err := ValidateMessageVote(userState, voteMessage); err == nil {
			t.FailNow()
		}
	})
	t.Run("HostState_JOINING valid", func(t *testing.T) {
		userState.HostState = constants.HostState_JOINING
		if err := ValidateMessageVote(userState, voteMessage); err != nil {
			log.Fatal(err)
		}
	})
	t.Run("HostState_HOSTING valid", func(t *testing.T) {
		userState.HostState = constants.HostState_HOSTING
		if err := ValidateMessageVote(userState, voteMessage); err != nil {
			log.Fatal(err)
		}
	})

}
