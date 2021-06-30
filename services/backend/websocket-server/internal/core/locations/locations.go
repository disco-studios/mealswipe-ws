package locations

import (
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GrabNextForUser(userState *users.UserState) (loc *mealswipepb.Location, err error) {
	ind, err := business.DbGameNextVoteInd(userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, err = business.DbLocationFromInd(userState.JoinedSessionId, int64(ind))
	return
}

func SendNextToUser(userState *users.UserState) (err error) {
	loc, err := GrabNextForUser(userState)
	if err != nil {
		return
	}

	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		Location: loc,
	})
	return
}
