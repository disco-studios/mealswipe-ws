package codes

import (
	"context"
	"math/rand"
	"time"

	"go.elastic.co/apm"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/msredis"
)

type CodeAlreadyExistsError struct {
}

func (e *CodeAlreadyExistsError) Error() string {
	return "code already claimed"
}

func attemptReserveCode(ctx context.Context, sessionId string, code string) (err error) {
	span, ctx := apm.StartSpan(ctx, "attemptReserveCode", "codes")
	defer span.End()

	// TODO Handle this a bit better, we could miss errors
	res, err := msredis.SetNX(ctx, keys.BuildCodeKey(code), sessionId, time.Hour*24).Result()
	if !res {
		// TODO This can probably be done better
		return &CodeAlreadyExistsError{}
	}
	return
}

func generateRandomRaw() int {
	randSource := rand.NewSource(time.Now().UnixNano())
	return rand.New(randSource).Intn(MAX_SESSION_CODE_RAW)
}

func encodeRaw(rawCode int) string {
	out := ""
	for i := 0; i < SESSION_CODE_LENGTH; i++ {
		out = SESSION_CODE_CHARSET[rawCode%SESSION_CODE_BASE] + out
		rawCode = rawCode / SESSION_CODE_BASE
	}
	return out
}

func decodeRaw(code string) int {
	out := 0
	// Go through each digit
	for _, codeChar := range code {
		out *= SESSION_CODE_BASE
		out += findValueOfCodeChar(string(codeChar))
	}
	return out
}

func findValueOfCodeChar(codeChar string) int {
	for charInd, charsetChar := range SESSION_CODE_CHARSET {
		if charsetChar == codeChar {
			return charInd
		}
	}
	panic("Could not find char for int")
}
