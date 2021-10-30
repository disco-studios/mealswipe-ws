package codes

import (
	"context"
	"fmt"
	"math"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
)

const MAX_CODE_ATTEMPTS int = 6 // 1-(1000000/(21^6))^6 = 0.999999999, aka almost certain with 1mil codes/day
// vowels removed to reduce odds of bad words
var SESSION_CODE_CHARSET []string = []string{
	"B", "C", "D", "F", "G", "H", "J",
	"K", "L", "M", "N", "P", "Q", "R",
	"S", "T", "V", "W", "X", "Y", "Z",
}

const SESSION_CODE_LENGTH int = 6

var SESSION_CODE_BASE = len(SESSION_CODE_CHARSET)
var MAX_SESSION_CODE_RAW int = int(math.Pow(
	float64(SESSION_CODE_BASE),
	float64(SESSION_CODE_LENGTH),
))

func Reserve(ctx context.Context, sessionId string) (code string, err error) {
	logger := logging.Get()
	for i := 0; i < MAX_CODE_ATTEMPTS; i++ {
		code = encodeRaw(generateRandomRaw())
		err = attemptReserveCode(ctx, sessionId, code)
		if err == nil { // TODO Handle errors other than the one we made
			logger.Info("reserved code", logging.Metric("code_collision"), zap.Bool("collision", false), zap.Error(err))
			return
		} else {
			logger.Info("failed to reserve code", logging.Metric("code_collision"), zap.Bool("collision", true), zap.Error(err))
		}
	}
	err = fmt.Errorf("ran out of attempts: %w", err)
	return
}
