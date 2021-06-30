package sessions

import (
	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core/codes"
	"mealswipe.app/mealswipe/internal/core/users"
)

const MAX_CODE_ATTEMPTS int = 6 // 1-(1000000/(21^6))^6 = 0.999999999, aka almost certain with 1mil codes/day

func Create(userState *users.UserState) (sessionID string, code string, err error) {
	sessionID = "s-" + uuid.NewString()
	code, err = reserveSessionCode(sessionID)
	if err != nil {
		return
	}
	err = business.DbSessionCreate(code, sessionID, userState.UserId)
	return
}

func reserveSessionCode(sessionId string) (code string, err error) {
	for i := 0; i < MAX_CODE_ATTEMPTS; i++ {
		code = codes.EncodeRaw(codes.GenerateRandomRaw())
		err = business.DbCodeReserve(sessionId, code)
		if err == nil {
			return
		}
	}
	panic("Ran out of tries")
}

func JoinById(userState *users.UserState, sessionId string, code string) (err error) {
	redisPubsub, err := business.DbSessionJoinById(userState.UserId, sessionId, userState.Nickname, userState.PubsubChannel)
	if err != nil {
		return
	}

	userState.RedisPubsub = redisPubsub
	userState.JoinedSessionId = sessionId
	userState.JoinedSessionCode = code

	return
}
