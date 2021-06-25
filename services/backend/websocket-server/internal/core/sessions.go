package core

import (
	"log"

	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/business"
)

const MAX_CODE_ATTEMPTS int = 6 // 1-(1000000/(21^6))^6 = 0.999999999, aka almost certain with 1mil codes/day

func CreateSession(userState *UserState) (sessionID string, code string, err error) {
	sessionID = "s-" + uuid.NewString()
	code, err = reserveSessionCode(sessionID)
	if err != nil {
		return
	}
	err = business.DbCreateSession(code, sessionID, userState.UserId)
	return
}

func reserveSessionCode(sessionId string) (code string, err error) {
	for i := 0; i < MAX_CODE_ATTEMPTS; i++ {
		code = EncodeRawCode(GenerateRandomRawCode())
		err = business.ReserveCode(sessionId, code)
		if err == nil {
			if i > 0 {
				log.Println(i)
			}
			return
		}
	}
	panic("Ran out of tries")
}
