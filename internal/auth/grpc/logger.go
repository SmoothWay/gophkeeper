package grpcapp

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type rpcLogger struct {
	log *slog.Logger
}

func (l *rpcLogger) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	l.log.Log(ctx, slog.Level(slog.LevelInfo), msg, fields...)
}
