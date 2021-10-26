package start

import (
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageStart(userState *users.UserState, startMessage *mealswipepb.StartMessage) (err error) {
	err = database.Start(userState.JoinedSessionCode, userState.JoinedSessionId, startMessage.Lat, startMessage.Lng, startMessage.Radius, startMessage.CategoryId)
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
