package auth

import (
	"context"
	authv1 "github.com/Fitnow08/fitnow-proto/pkg/gen/go/v1/auth"
	"google.golang.org/grpc"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
}

func RegisterGrpcServer(server *grpc.Server) {
	authv1.RegisterAuthServiceServer(server, &Handler{})
}

func (s *Handler) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Handler) Register(ctx context.Context, request *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Handler) NewTokens(ctx context.Context, request *authv1.NewTokensRequest) (*authv1.NewTokensResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Handler) VerifyAccount(ctx context.Context, request *authv1.VerifyAccountRequest) (*authv1.VerifyAccountResponse, error) {
	//TODO implement me
	panic("implement me")
}
