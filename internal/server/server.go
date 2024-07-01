package server

import (
	"log/slog"
	"net/http"

	"github.com/SmoothWay/gophkeeper/internal/server/clients"
	"github.com/SmoothWay/gophkeeper/internal/server/config"
	"github.com/SmoothWay/gophkeeper/internal/server/handler"
	"github.com/SmoothWay/gophkeeper/internal/server/service"
	"github.com/SmoothWay/gophkeeper/internal/server/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	db *pgxpool.Pool
}

func Run(log *slog.Logger, cfg *config.Config) {
	// TODO gracefull shutdown
	db, err := storage.New(cfg.DatabaseURL, cfg.QueryTimeout)
	if err != nil {
		panic(err)
	}
	storageKeeper := storage.NewKeeperPostgres(db, cfg.QueryTimeout)
	serviceKeeper := service.New(log, storageKeeper, cfg.Key)
	conns := clients.NewWSConnMap()
	h := handler.NewHandler(log, serviceKeeper, conns)

	http.HandleFunc("/ws", h.Handle)

	err = http.ListenAndServeTLS(cfg.WS.Address, cfg.CertFile, cfg.KeyFile, nil)
	if err != nil {
		panic(err)
	}
}
