package handlers

import (
	"mealswipe.app/mealswipe/internal/core/game"
	"mealswipe.app/mealswipe/internal/core/locations"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *users.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	err = game.Vote(userState.UserId, int64(voteMessage.Index), voteMessage.Vote)
	if err != nil {
		return
	}

	go game.CheckWin(userState) // TODO This could throw an error, figure out how to handle

	err = locations.SendNextToUser(userState)

	return err
}
