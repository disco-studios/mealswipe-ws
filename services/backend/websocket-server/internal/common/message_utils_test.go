package common

import (
	"testing"

	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HasCreateMessageTest(t *testing.T) {
	websocketMessage := &mealswipepb.WebsocketMessage{}

	t.Run("fails without message", func(t *testing.T) {
		if HasCreateMessage(websocketMessage) {
			t.FailNow()
		}
	})

	t.Run("succeeds with message", func(t *testing.T) {
		websocketMessage.CreateMessage = &mealswipepb.CreateMessage{}

		if !HasCreateMessage(websocketMessage) {
			t.FailNow()
		}
	})
}

func HasStartMessageTest(t *testing.T) {
	websocketMessage := &mealswipepb.WebsocketMessage{}

	t.Run("fails without message", func(t *testing.T) {
		if HasStartMessage(websocketMessage) {
			t.FailNow()
		}
	})

	t.Run("succeeds with message", func(t *testing.T) {
		websocketMessage.StartMessage = &mealswipepb.StartMessage{}

		if !HasStartMessage(websocketMessage) {
			t.FailNow()
		}
	})
}

func HasJoinMessageTest(t *testing.T) {
	websocketMessage := &mealswipepb.WebsocketMessage{}

	t.Run("fails without message", func(t *testing.T) {
		if HasJoinMessage(websocketMessage) {
			t.FailNow()
		}
	})

	t.Run("succeeds with message", func(t *testing.T) {
		websocketMessage.JoinMessage = &mealswipepb.JoinMessage{}

		if !HasJoinMessage(websocketMessage) {
			t.FailNow()
		}
	})
}

func HasVoteMessageTest(t *testing.T) {
	websocketMessage := &mealswipepb.WebsocketMessage{}

	t.Run("fails without message", func(t *testing.T) {
		if HasVoteMessage(websocketMessage) {
			t.FailNow()
		}
	})

	t.Run("succeeds with message", func(t *testing.T) {
		websocketMessage.VoteMessage = &mealswipepb.VoteMessage{}

		if !HasVoteMessage(websocketMessage) {
			t.FailNow()
		}
	})
}
