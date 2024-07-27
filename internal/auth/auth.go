package auth

import (
	"log/slog"

	"github.com/SmoothWay/gophkeeper/internal/auth/config"
	grpcApp "github.com/SmoothWay/gophkeeper/internal/auth/grpc"
	"github.com/SmoothWay/gophkeeper/internal/auth/service"
	"github.com/SmoothWay/gophkeeper/internal/auth/storage"
)

type App struct {
	grpcApp *grpcApp.App
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	err := storage.Migrate(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	userStorage, err := storage.NewUser(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	appStorage, err := storage.NewApp(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	authService := service.New(log, userStorage, appStorage, cfg.TokenTTL)

	grpcApp, err := grpcApp.New(log, authService, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		grpcApp: grpcApp,
	}, nil
}

func (app *App) MustRun() {
	app.grpcApp.MustRun()
}

func (app *App) Stop() {
	app.grpcApp.Stop()
}
