package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fitnow08/fitnow-backend-sso/internal/feature/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
)

type GRPCServer struct {
	log        *slog.Logger
	GRPCServer *grpc.Server
	port       string
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func NewGRPCServer(log *slog.Logger, grpcPort string, service auth.AuthService) (*GRPCServer, error) {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandlerContext(func(ctx context.Context, p any) error {
			log.Error("recovered from panic", "panic", p)
			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
		),
	)
	auth.RegisterGrpcServer(gRPCServer, service, log)
	reflection.Register(gRPCServer)
	return &GRPCServer{GRPCServer: gRPCServer, log: log, port: grpcPort}, nil
}
func (a *GRPCServer) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
func (a *GRPCServer) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With("op", op, "port", a.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", a.port))
	if err != nil {
		log.Error("failed to listen", "err", err)
		return errors.New("failed init grpc app")
	}
	if err := a.GRPCServer.Serve(l); err != nil {
		log.Error("failed to serve", "err", err)
		return errors.New("failed init grpc app")
	}
	return nil
}

func (a *GRPCServer) Stop() {
	const op = "grpcapp.Stop"
	log := a.log.With("op", op, "port", a.port)
	log.Info("stopping grpc server")
	a.GRPCServer.GracefulStop()
}
