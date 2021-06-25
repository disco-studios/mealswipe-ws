package core

import "mealswipe.app/mealswipe/internal/business"

func UserJoinSessionById(userState *UserState, sessionId string, code string) (err error) {
	redisPubsub, err := business.DbUserJoinSessionById(userState.UserId, sessionId, userState.Nickname, userState.PubsubChannel)
	if err != nil {
		return
	}

	userState.RedisPubsub = redisPubsub
	userState.JoinedSessionId = sessionId
	userState.JoinedSessionCode = code

	return
}
