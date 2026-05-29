package app

import (
	"context"
	"fmt"
	"github.com/Fitnow08/fitnow-backend-sso/internal/config"
	grpcserver "github.com/Fitnow08/fitnow-backend-sso/internal/servers/grpc"
	"github.com/Fitnow08/fitnow-backend-sso/pkg/logger"
	"log/slog"
)

type App struct {
	Log          *slog.Logger
	Cfg          *config.Config
	GRPCServer   *grpcserver.GRPCServer
	DB           *Database
	CancelLogger func()
}

func NewApp(ctx context.Context) (*App, error) {
	cfg := config.InitConfig()
	l, cancelogger := logger.SetupLogger(ctx, cfg.Env, fmt.Sprintf("%s:%s", "", ""))
	databases, err := NewDataBases(cfg, l)
	if err != nil {
		return nil, err
	}

	repo := NewRepositories(l, databases)
	srv := NewServices(l, cfg, repo)
	GRPCServer, err := grpcserver.NewGRPCServer(l, cfg.GRPC.Port, srv.AuthService)
	if err != nil {
		return nil, err
	}
	return &App{
		Log:          l,
		Cfg:          cfg,
		CancelLogger: cancelogger,
		GRPCServer:   GRPCServer,
		DB:           databases,
	}, nil
}
