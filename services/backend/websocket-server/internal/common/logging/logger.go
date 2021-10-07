package logging

import (
	"go.uber.org/zap"
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
