package vote

import (
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common/logging"
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageVote(userState *users.UserState, voteMessage *mealswipepb.VoteMessage) (err error) {
	logger := logging.Get()
	err = database.Vote(userState.UserId, userState.JoinedSessionId, voteMessage.Index, voteMessage.Vote)
	if err != nil {
		return
	}

	logger.Info("user_vote",
		logging.Metric("swipe_dir"),
		zap.Bool("right", voteMessage.Vote),
		logging.SessionId(userState.JoinedSessionId),
		zap.Int32("index", voteMessage.Index),
		logging.UserId(userState.UserId),
	)
	go database.CheckWin(userState) // TODO This could throw an error, figure out how to handle

	err = database.SendNextLocToUser(userState)

	return err
}
