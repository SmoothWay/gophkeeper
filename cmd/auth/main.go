package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SmoothWay/gophkeeper/internal/auth"
	"github.com/SmoothWay/gophkeeper/internal/auth/config"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.MustLoad()
	log.Debug("starting application", slog.Any("config", cfg))

	app, err := auth.New(log, cfg)
	if err != nil {
		log.Error(
			"error creating application service",
			logger.Err(err),
		)
		return
	}
	go app.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))
	app.Stop()
	log.Info("application stopped!")
}
