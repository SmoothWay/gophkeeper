package client

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/SmoothWay/gophkeeper/internal/client/config"
	"github.com/SmoothWay/gophkeeper/internal/client/storage"
	"github.com/SmoothWay/gophkeeper/internal/grpcclient"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
)

type AppClient struct {
	grpcClient   *grpcclient.GRPCClient
	log          *slog.Logger
	queryTimeout time.Duration
	storagePath  string
	grpcAddress  string
	WSURL        string
	caCertFile   string
}

func NewAppClient(log *slog.Logger, cfg *config.ClientConfig) *AppClient {
	return &AppClient{
		log:          log,
		storagePath:  cfg.StoragePath,
		grpcAddress:  cfg.GRPCAddress,
		WSURL:        cfg.WSURL,
		queryTimeout: cfg.QueryTime,
		caCertFile:   cfg.CaCertFile,
	}
}

func (app *AppClient) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"

	log := app.log.With(
		slog.String("op", op),
	)

	// app.ch = make(chan models.Message)

	err := storage.Migrate(app.storagePath)
	if err != nil {
		log.Error(
			"migration database error",
			logger.Err(err),
		)
		stop <- syscall.SIGTERM
		return
	}

	dbCred, err := storage.NewCredentials(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init credentials storage")
		stop <- syscall.SIGTERM
		return
	}

	dbText, err := storage.NewText(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init text storage")
		stop <- syscall.SIGTERM
		return
	}

	dbBin, err := storage.NewBinary(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init binary storage")
		stop <- syscall.SIGTERM
		return
	}

	dbCard, err := storage.NewCard(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init card storage")
		stop <- syscall.SIGTERM
		return
	}

	app.keeper = service.NewKeeper()
}
