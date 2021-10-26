package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func SetLogger(logr *zap.Logger) {
	if logger == nil {
		logger = logr
	} else {
		logger.Warn("tried to set logger when we already have one")
	}
}

func Get() *zap.Logger {
	return logger
}

func SessionId(sessionId string) zapcore.Field {
	return zap.String("session_id", sessionId)
}

func UserId(userId string) zapcore.Field {
	return zap.String("user_id", userId)
}

func LocId(locId string) zapcore.Field {
	return zap.String("loc_id", locId)
}

func LocName(locId string) zapcore.Field {
	return zap.String("loc_name", locId)
}

func Metric(metic string) zapcore.Field {
	return zap.String("metric", metic)
}

func Code(code string) zapcore.Field {
	return zap.String("code", code)
}
