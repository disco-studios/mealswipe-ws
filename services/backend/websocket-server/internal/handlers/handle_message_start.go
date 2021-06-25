package handlers

import (
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
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
