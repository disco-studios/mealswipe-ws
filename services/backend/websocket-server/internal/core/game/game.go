package game

import (
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/logging"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func CheckWin(userState *users.UserState) (err error) {
	logger := logging.Get()
	win, winIndex, err := business.DbGameCheckWin(userState.JoinedSessionId)
	if err != nil {
		logger.Error("failed to check for win", logging.SessionId(userState.JoinedSessionId))
		return
	}

	if win {
		var loc *mealswipepb.Location
		loc, err = business.DbLocationFromInd(userState.JoinedSessionId, winIndex)
		if err != nil {
			return
		}

		err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
			GameWinMessage: &mealswipepb.GameWinMessage{
				Locations: []*mealswipepb.WinningLocation{
					{
						Location: loc,
						Votes:    0, // TODO: Impl
					},
				},
			},
		})
		if err != nil {
			return
		}
	}
	return
}

func Vote(userId string, sessionId string, index int32, state bool) (err error) {
	return business.DbGameSendVote(userId, sessionId, index, state)
}
