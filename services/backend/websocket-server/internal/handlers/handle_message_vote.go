package handlers

import (
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common/logging"
	"mealswipe.app/mealswipe/internal/core/game"
	"mealswipe.app/mealswipe/internal/core/locations"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *users.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	logger := logging.Get()
	err = game.Vote(userState.UserId, userState.JoinedSessionId, voteMessage.Index, voteMessage.Vote)
	if err != nil {
		return
	}

	logger.Info("user_vote",
		logging.Metric("swipe_dir"),
		zap.Bool("right", voteMessage.Vote),
		logging.LocId(userState.JoinedSessionId),
		zap.Int32("index", voteMessage.Index),
		logging.UserId(userState.UserId),
	)
	go game.CheckWin(userState) // TODO This could throw an error, figure out how to handle

	err = locations.SendNextToUser(userState)

	return err
}
