package app

import (
	"github.com/Fitnow08/fitnow-backend-sso/internal/feature/auth"
	"log/slog"
)

type Repositories struct {
	AuthRepository   *auth.Repository
	VerifyRepository *auth.VerifyRepository
}

func NewRepositories(l *slog.Logger, db *Database) *Repositories {
	return &Repositories{
		AuthRepository:   auth.NewRepository(l, db.PrimaryDB),
		VerifyRepository: auth.NewVerifyRepository(l, db.RedisDB),
	}
}
