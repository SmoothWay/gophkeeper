package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func init() {
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log = slogger
}

func NewLogger() *slog.Logger {
	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
