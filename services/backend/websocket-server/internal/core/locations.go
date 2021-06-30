package core

import (
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GrabNextLocForUser(userState *UserState) (loc *mealswipepb.Location, err error) {
	ind, err := business.DbGameNextVoteInd(userState.JoinedSessionId, userState.UserId)
	if err != nil {
		return
	}

	loc, err = business.DbLocationFromInd(userState.JoinedSessionId, int64(ind))
	return
}

func SendNextLocToUser(userState *UserState) (err error) {
	loc, err := GrabNextLocForUser(userState)
	if err != nil {
		return
	}

	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		Location: loc,
	})
	return
}
