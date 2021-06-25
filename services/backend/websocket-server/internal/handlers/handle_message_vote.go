package handlers

import (
	"errors"

	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *core.UserState, endMessage *mealswipepb.VoteMessage) (err error) {
	return errors.New("unimplemented")
}
