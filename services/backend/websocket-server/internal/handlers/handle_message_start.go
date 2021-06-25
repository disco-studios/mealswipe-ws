package handlers

import (
	"mealswipe.app/mealswipe/business"
	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
)

func HandleMessageStart(userState *core.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	err = business.DbStartSession(userState.JoinedSessionCode, userState.JoinedSessionId, startMessage.Lat, startMessage.Lng)
	if err != nil {
		return
	}

	err = userState.SendPubsubMessage("start")
	if err != nil {
		return
	}
	return
}
