package handlers

import (
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *core.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	err = business.DbVote(userState.UserId, int64(voteMessage.Index), voteMessage.Vote)

	go core.CheckWin(userState) // TODO This could throw an error, figure out how to handle

	return err
}
