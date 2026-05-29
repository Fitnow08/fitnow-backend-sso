package app

import (
	"github.com/Fitnow08/fitnow-backend-sso/internal/config"
	"github.com/Fitnow08/fitnow-backend-sso/internal/feature/auth"
	"log/slog"
)

type Services struct {
	AuthService *auth.Service
}

func NewServices(l *slog.Logger, cfg *config.Config, repo *Repositories) *Services {
	mailer := auth.NewMailer(l, cfg.Mail)
	return &Services{
		AuthService: auth.NewService(l, repo.AuthRepository, mailer, repo.VerifyRepository),
	}
}
