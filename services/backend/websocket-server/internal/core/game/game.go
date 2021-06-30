package game

import (
	"log"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func CheckWin(userState *users.UserState) (err error) {
	win, winIndex, err := business.DbGameCheckWin(userState.JoinedSessionId)
	if err != nil {
		log.Println(err)
		return
	}

	if win {
		var loc *mealswipepb.Location
		loc, err = business.DbLocationFromInd(userState.JoinedSessionId, int64(winIndex))
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

func Vote(userId string, index int64, state bool) (err error) {
	return business.DbGameSendVote(userId, index, state)
}
