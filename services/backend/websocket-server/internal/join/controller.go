package join

import (
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
)

func JoinSessionById(userState *users.UserState, sessionId string, code string) (err error) {
	redisPubsub, err := database.JoinById(userState.UserId, sessionId, userState.Nickname, userState.PubsubChannel)
	if err != nil {
		return
	}

	userState.RedisPubsub = redisPubsub
	userState.JoinedSessionId = sessionId
	userState.JoinedSessionCode = code

	return
}
