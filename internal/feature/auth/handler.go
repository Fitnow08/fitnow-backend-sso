package auth

import (
	"context"
	"github.com/Fitnow08/fitnow-backend-sso/internal/models/domain"
	authv1 "github.com/Fitnow08/fitnow-proto/pkg/gen/go/v1/auth"
	"google.golang.org/grpc"
	"log/slog"
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*UserDB, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GenerateNewTokens(ctx context.Context, token string) (*Tokens, error)
	VerifyAccount(ctx context.Context, email string, code int) (*domain.User, error)
	ResendCode(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email string) error
	ConfirmResetPassword(ctx context.Context, email string, newPassword string, code int) error
}
type Handler struct {
	authv1.UnimplementedAuthServiceServer
	service AuthService
	log     *slog.Logger
}

func RegisterGrpcServer(server *grpc.Server, service AuthService, log *slog.Logger) {
	authv1.RegisterAuthServiceServer(server, &Handler{service: service, log: log})
}

func (h *Handler) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	const op = "Auth.Handler.Login"
	log := h.log.With(slog.String("op", op))

	user, err := h.service.Login(ctx, request.Email, request.Password)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &authv1.LoginResponse{
		Email:        user.Email,
		Title:        user.Title,
		Id:           user.ID.String(),
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}, nil
}

func (h *Handler) Register(ctx context.Context, request *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	const op = "Auth.Handler.Register"
	log := h.log.With(slog.String("op", op))
	_, err := h.service.Register(ctx, RegisterRequest{Email: request.Email, Password: request.Password, Name: request.Name})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.RegisterResponse{
		Ok: "ok",
	}, nil
}

func (h *Handler) NewTokens(ctx context.Context, request *authv1.NewTokensRequest) (*authv1.NewTokensResponse, error) {
	const op = "Auth.Handler.Login"
	log := h.log.With(slog.String("op", op))

	tokens, err := h.service.GenerateNewTokens(ctx, request.RefreshToken)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.NewTokensResponse{
		RefreshToken: tokens.RefreshToken,
		AccessToken:  tokens.AccessToken,
	}, nil
}

func (h *Handler) VerifyAccount(ctx context.Context, request *authv1.VerifyAccountRequest) (*authv1.VerifyAccountResponse, error) {
	const op = "Auth.Handler.Login"
	log := h.log.With(slog.String("op", op))

	varifyac, err := h.service.VerifyAccount(ctx, request.Email, int(request.VerifyCode))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.VerifyAccountResponse{
		Email:        varifyac.Email,
		Title:        varifyac.Title,
		Id:           varifyac.ID.String(),
		AccessToken:  varifyac.AccessToken,
		RefreshToken: varifyac.RefreshToken,
	}, nil
}
func (h *Handler) ResendVerifyCode(ctx context.Context, request *authv1.ResendVerifyCodeRequest) (*authv1.ResendVerifyCodeResponse, error) {
	const op = "Auth.Handler.ResendVerifyCode"
	log := h.log.With(slog.String("op", op))
	if err := h.service.ResendCode(ctx, request.Email); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.ResendVerifyCodeResponse{Ok: true}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, request *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	const op = "Auth.Handler.ResetPassword"
	log := h.log.With(slog.String("op", op))
	if err := h.service.ResetPassword(ctx, request.Email); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.ResetPasswordResponse{Ok: true}, nil
}
func (h *Handler) ConfirmResetPassword(ctx context.Context, request *authv1.ConfirmResetPasswordRequest) (*authv1.ConfirmResetPasswordResponse, error) {
	const op = "Auth.Handler.ConfirmResetPassword"
	log := h.log.With(slog.String("op", op))
	if err := h.service.ConfirmResetPassword(ctx, request.Email, request.NewPassword, int(request.Code)); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &authv1.ConfirmResetPasswordResponse{Ok: true}, nil
}
