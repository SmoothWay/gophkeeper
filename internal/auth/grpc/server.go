package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/SmoothWay/gophkeeper/internal/auth/config"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	service    Auth
	port       int
}

func New(log *slog.Logger, authService Auth, cfg *config.Config) (*App, error) {
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(&rpcLogger{log: log})),
		grpc.Creds(insecure.NewCredentials()),
	)

	Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		service:    authService,
		port:       cfg.GRPC.Port,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		a.log.Error(
			"error running GRPC server",
			logger.Err(err),
		)
		return
	}
}

func (a *App) Run() error {
	const op = "auth.grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("gRPC server is running", slog.String("addr", listener.Addr().String()))

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Stop graceful stopped GRPC server
func (a *App) Stop() {
	const op = "auth.grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
	a.service.Close()
}
