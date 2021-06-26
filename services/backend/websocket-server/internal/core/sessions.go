package core

import (
	"log"

	"github.com/google/uuid"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
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

func CheckWin(userState *UserState) (err error) {
	win, winIndex, err := business.DbCheckWin(userState.JoinedSessionId)
	if err != nil {
		log.Println(err)
		return
	}

	if win {
		var loc *mealswipepb.Location
		loc, err = business.DbGrabLocationFromInd(userState.JoinedSessionId, int64(winIndex))
		if err != nil {
			return
		}

		err = userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
			GameWinMessage: &mealswipepb.GameWinMessage{
				Locations: []*mealswipepb.WinningLocation{
					{
						Location: loc,
						Votes:    0, // TODO: Impl
					},
				},
			},
		})
		if err != nil {
			return
		}
	}
	return
}

func reserveSessionCode(sessionId string) (code string, err error) {
	for i := 0; i < MAX_CODE_ATTEMPTS; i++ {
		code = EncodeRawCode(GenerateRandomRawCode())
		err = business.ReserveCode(sessionId, code)
		if err == nil {
			return
		}
	}
	panic("Ran out of tries")
}
