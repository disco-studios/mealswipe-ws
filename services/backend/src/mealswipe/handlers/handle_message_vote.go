package handlers

import (
	"errors"

	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *core.UserState, endMessage *mealswipepb.VoteMessage) (err error) {
	return errors.New("unimplemented")
}
