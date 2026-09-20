package queue

import (
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/logger"
)

type asynqSlogLogger struct{}

func NewAsynqLogger() asynq.Logger {
	return asynqSlogLogger{}
}

func (asynqSlogLogger) Debug(args ...any) {
	slog.Debug(fmt.Sprint(args...))
}

func (asynqSlogLogger) Info(args ...any) {
	slog.Info(fmt.Sprint(args...))
}

func (asynqSlogLogger) Warn(args ...any) {
	slog.Warn(fmt.Sprint(args...))
}

func (asynqSlogLogger) Error(args ...any) {
	slog.Error(fmt.Sprint(args...))
}

func (asynqSlogLogger) Fatal(args ...any) {
	logger.Fatalf("%s", fmt.Sprint(args...))
}
