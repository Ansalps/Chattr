package utils

import (
	"context"

	"github.com/Ansalps/Chattr_Post_Relation_Service/infrastructure/logger"
)

type contextKey string

const loggerKey contextKey = "logger"

func SetLogger(ctx context.Context, log logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func GetLogger(ctx context.Context) logger.Logger {
	log, ok := ctx.Value(loggerKey).(logger.Logger)
	if !ok {
		panic("logger not found in context")
	}
	return log
}
