package validators

import (
	"log"
	"testing"

	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO Make sure errors are of right type
func TestValidateMessage(t *testing.T) {
	t.Run("valid create message", func(t *testing.T) {
		userState := core.CreateUserState()
		createMessage := &mealswipepb.WebsocketMessage{
			CreateMessage: &mealswipepb.CreateMessage{
				Nickname: "Cam the Man",
			},
		}

		err := ValidateMessage(userState, createMessage)
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("valid join message", func(t *testing.T) {
		userState := core.CreateUserState()
		joinMessage := &mealswipepb.WebsocketMessage{
			JoinMessage: &mealswipepb.JoinMessage{
				Nickname: "Cam the Man",
				Code:     "ABCDEF",
			},
		}

		err := ValidateMessage(userState, joinMessage)
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("valid start message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_HOSTING
		startMessage := &mealswipepb.WebsocketMessage{
			StartMessage: &mealswipepb.StartMessage{
				Lat: 0.0,
				Lng: 0.0,
			},
		}

		err := ValidateMessage(userState, startMessage)
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("valid vote message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_HOSTING
		startMessage := &mealswipepb.WebsocketMessage{
			VoteMessage: &mealswipepb.VoteMessage{
				Index: 0,
				Vote:  true,
			},
		}

		err := ValidateMessage(userState, startMessage)
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("invalid HostState create message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_HOSTING
		createMessage := &mealswipepb.WebsocketMessage{
			CreateMessage: &mealswipepb.CreateMessage{
				Nickname: "Cam the Man",
			},
		}

		err := ValidateMessage(userState, createMessage)
		if err == nil {
			t.FailNow()
		}
	})

	t.Run("invalid HostState join message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_HOSTING
		joinMessage := &mealswipepb.WebsocketMessage{
			JoinMessage: &mealswipepb.JoinMessage{
				Nickname: "Cam the Man",
				Code:     "ABCDEF",
			},
		}

		err := ValidateMessage(userState, joinMessage)
		if err == nil {
			t.FailNow()
		}
	})

	t.Run("invalid HostState start message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_JOINING
		startMessage := &mealswipepb.WebsocketMessage{
			StartMessage: &mealswipepb.StartMessage{
				Lat: 0.0,
				Lng: 0.0,
			},
		}

		err := ValidateMessage(userState, startMessage)
		if err == nil {
			t.FailNow()
		}
	})

	t.Run("invalid HostState vote message", func(t *testing.T) {
		userState := core.CreateUserState()
		userState.HostState = core.HostState_UNIDENTIFIED
		startMessage := &mealswipepb.WebsocketMessage{
			VoteMessage: &mealswipepb.VoteMessage{
				Index: 0,
				Vote:  true,
			},
		}

		err := ValidateMessage(userState, startMessage)
		if err == nil {
			t.FailNow()
		}
	})

	t.Run("invalid empty message", func(t *testing.T) {
		userState := core.CreateUserState()
		emptyMessage := &mealswipepb.WebsocketMessage{}

		err := ValidateMessage(userState, emptyMessage)
		if err == nil {
			t.FailNow()
		}
	})

}
