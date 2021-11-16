package logging

import (
	"context"

	"go.elastic.co/apm/module/apmzap"
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

func metric(logga *zap.Logger, metric_str string) *zap.Logger {
	return logga.With(zap.Namespace("metric"), zap.String("type", metric_str), zap.Namespace(metric_str))
}

func Metric(metric_str string) *zap.Logger {
	return metric(logger, metric_str)
}

func MetricCtx(ctxt context.Context, metric_str string) *zap.Logger {
	return metric(ctx(logger, ctxt), metric_str)
}

func ctx(logga *zap.Logger, ctx context.Context) *zap.Logger {
	if val, ok := ctx.Value("user.id").(string); ok {
		logga = logga.With(
			zap.String("user.id", val),
		)
	}

	// if val, ok := ctx.Value("host.state").(string); ok {
	// 	logga = logga.With(
	// 		zap.String("host.state", val),
	// 	)
	// }

	if val, ok := ctx.Value("session.id").(string); ok {
		logga = logga.With(
			zap.String("session.id", val),
		)
	}

	return ApmCtx(ctx, logga)
}

func Ctx(ctxt context.Context) *zap.Logger {
	return ctx(logger, ctxt)
}

func ApmCtx(ctx context.Context, logga *zap.Logger) *zap.Logger {
	traceContextFields := apmzap.TraceContext(ctx)

	return logga.With(traceContextFields...)
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

func Code(code string) zapcore.Field {
	return zap.String("code", code)
}
