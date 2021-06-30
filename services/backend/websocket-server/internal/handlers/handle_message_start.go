package handlers

import (
	"mealswipe.app/mealswipe/internal/core/sessions"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageStart(userState *users.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	err = sessions.Start(userState.JoinedSessionCode, userState.JoinedSessionId, startMessage.Lat, startMessage.Lng)
	if err != nil {
		return
	}

	err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
		GameStartedMessage: &mealswipepb.GameStartedMessage{},
	})
	if err != nil {
		return
	}
	return
}
