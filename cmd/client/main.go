package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SmoothWay/gophkeeper/internal/client"
	"github.com/SmoothWay/gophkeeper/internal/client/config"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	log := logger.NewLogger()
	cfg := config.MustLoad()
	log.Debug("starting client application", slog.Any("config", cfg))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := client.NewAppClient(log, cfg)
	stop := make(chan os.Signal, 1)
	go app.Run(ctx, stop)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sign := <-stop
	log.Debug("stopping application", slog.String("signal", sign.String()))

	app.Stop()

	log.Debug("application stopped")
}
