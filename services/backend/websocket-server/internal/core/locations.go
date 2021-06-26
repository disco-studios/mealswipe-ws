package core

import (
	"log"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func GrabNextLocForUser(userState *UserState) (loc *mealswipepb.Location, err error) {
	ind, err := business.DbGrabAndIncrVoteInd(userState.JoinedSessionId, userState.UserId)
	log.Println("ind:", ind)
	if err != nil {
		return
	}

	loc, err = business.DbGrabLocationFromInd(userState.JoinedSessionId, int64(ind))
	log.Println("loc:", loc)
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
