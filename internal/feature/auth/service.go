package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fitnow08/fitnow-backend-sso/internal/models/domain"
	authv1 "github.com/Fitnow08/fitnow-proto/pkg/gen/go/v1/auth"
	"github.com/google/uuid"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"log/slog"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, email, title string, password []byte) (*UserDB, error)
	UserByEmail(ctx context.Context, email string) (*UserDB, error)
	SetVerifyAccount(ctx context.Context, email string) error
	UpdateUserPassword(ctx context.Context, email string, password []byte) error
	GetAllUsers(ctx context.Context) ([]*UserDB, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*UserDB, error)
}

type MailSender interface {
	SendVerifyCode(ctx context.Context, to string, code int) error
}

type VerifyCodeStorage interface {
	Save(ctx context.Context, email string, code int) error
	Get(ctx context.Context, email string) (int, error)
	Delete(ctx context.Context, email string) error
}

type Service struct {
	log            *slog.Logger
	authrepository AuthRepository
	mailer         MailSender
	verifyrepo     VerifyCodeStorage
}

func NewService(log *slog.Logger, authrepository AuthRepository, mailer MailSender, verifyrepo VerifyCodeStorage) *Service {
	return &Service{log: log, authrepository: authrepository, mailer: mailer, verifyrepo: verifyrepo}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserDB, error) {
	const op = "Auth.Service.Register"
	log := s.log.With("op", op)
	_, err := s.authrepository.UserByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hashpass, err := GeneratePasswordHash(req.Password)
	if err != nil {
		log.Error("failed to generate password hash", "error", err)
		return nil, err
	}
	userdb, err := s.authrepository.CreateUser(ctx, req.Email, req.Name, hashpass)
	if err != nil {
		log.Error("failed to create user", "error", err)
		return nil, err
	}

	code := rand.IntN(900_000) + 100_000
	if err := s.verifyrepo.Save(ctx, req.Email, code); err != nil {
		log.Error("failed to save verify code", "error", err)
		return nil, err
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.mailer.SendVerifyCode(sendCtx, req.Email, code); err != nil {
			s.log.Error("failed to send verify code", "error", err, "email", req.Email)
		}
	}()

	return &UserDB{
		ID:       userdb.ID,
		Email:    userdb.Email,
		Title:    userdb.Title,
		Password: []byte(""),
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*domain.User, error) {
	const op = "Auth.Service.Login"
	log := s.log.With("op", op)

	user, err := s.authrepository.UserByEmail(ctx, email)
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		return nil, err
	}
	if !VerifyPassword(user.Password, password) {
		return nil, fmt.Errorf("invalid password")
	}
	access, refresh, err := GenerateJwtTokens(user.ID, "user")
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return nil, err
	}
	return &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		Title:        user.Title,
		Role:         user.Role,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Service) GenerateNewTokens(ctx context.Context, token string) (*Tokens, error) {
	const op = "Auth.Service.GenerateNewTokens"
	log := s.log.With("op", op)
	log.Info("Generating new tokens")

	claims, err := ParseToken(token)
	if err != nil {
		log.Error("failed to parse refresh token", "error", err)
		return nil, err
	}

	access, refresh, err := GenerateJwtTokens(claims.ID, "user")
	if err != nil {
		log.Error("failed to generate tokens", "error", err)
		return nil, err
	}
	return &Tokens{
		RefreshToken: refresh,
		AccessToken:  access,
	}, nil
}

func (s *Service) VerifyAccount(ctx context.Context, email string, code int) (*domain.User, error) {
	const op = "Auth.Service.VerifyAccount"
	log := s.log.With("op", op, "email", email)

	saved, err := s.verifyrepo.Get(ctx, email)
	if err != nil {
		if errors.Is(err, ErrVerifyCodeNotFound) {
			return nil, ErrVerifyCodeNotFound
		}
		log.Error("failed to get verify code", "error", err)
		return nil, err
	}
	if saved != code {
		return nil, errors.New("invalid verify code")
	}

	user, err := s.authrepository.UserByEmail(ctx, email)
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		return nil, err
	}

	if err := s.authrepository.SetVerifyAccount(ctx, email); err != nil {
		log.Error("failed to set verify account", "error", err)
		return nil, err
	}
	if err := s.verifyrepo.Delete(ctx, email); err != nil {
		log.Warn("failed to delete verify code", "error", err)
	}

	access, refresh, err := GenerateJwtTokens(user.ID, "user")
	if err != nil {
		log.Error("failed to generate tokens", "error", err)
		return nil, err
	}
	return &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		Title:        user.Title,
		Role:         user.Role,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *Service) ResendCode(ctx context.Context, email string) error {
	const op = "Auth.Service.ResendCode"
	log := s.log.With("op", op)
	olduser, err := s.authrepository.UserByEmail(ctx, email)
	if err != nil {
		slog.Error("failed to get user by email", "error", err)
		return err
	}
	if olduser == nil {
		return errors.New("user not found")
	}
	if err := s.verifyrepo.Delete(ctx, email); err != nil {
		log.Error("failed to delete verify code", "error", err)
		return err
	}
	code := rand.IntN(900_000) + 100_000
	if err := s.verifyrepo.Save(ctx, email, code); err != nil {
		log.Error("failed to save verify code", "error", err)
		return err
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.mailer.SendVerifyCode(sendCtx, email, code); err != nil {
			log.Error("failed to send verify code", "error", err, "email", email)
		}
	}()
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, email string) error {
	const op = "Auth.Service.ResetPassword"

	log := s.log.With("op", op)

	_, err := s.authrepository.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user does not exist")
		}
		log.Error("failed to get user by email", "error", err)
		return err
	}
	code := rand.IntN(900_000) + 100_000
	if err := s.verifyrepo.Save(ctx, email, code); err != nil {
		log.Error("failed to save verify code", "error", err)
		return err
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.mailer.SendVerifyCode(sendCtx, email, code); err != nil {
			s.log.Error("failed to send verify code", "error", err, "email", email)
		}
	}()
	return nil
}

func (s *Service) ConfirmResetPassword(ctx context.Context, email string, newPassword string, code int) error {
	const op = "Auth.Service.ConfirmResetPassword"
	log := s.log.With("op", op)
	oldcode, err := s.verifyrepo.Get(ctx, email)
	if err != nil {
		log.Error("failed to get verify code", "error", err)
		return err
	}
	if oldcode != code {
		return errors.New("invalid verify code")
	}
	hashpass, err := GeneratePasswordHash(newPassword)
	if err != nil {
		log.Error("failed to generate password hash", "error", err)
		return err
	}
	if err := s.authrepository.UpdateUserPassword(ctx, email, hashpass); err != nil {
		log.Error("failed to update user password", "error", err)
		return err
	}
	if err := s.verifyrepo.Delete(ctx, email); err != nil {
		log.Error("failed to delete verify code", "error", err)
		return err
	}
	return nil
}

func (s *Service) GetAllUsers(ctx context.Context) ([]*authv1.User, error) {
	const op = "Auth.Service.GetAllUsers"
	log := s.log.With("op", op)
	users, err := s.authrepository.GetAllUsers(ctx)
	if err != nil {
		log.Error("failed to get all users", "error", err)
		return nil, err
	}
	newusers := make([]*authv1.User, 0, len(users))
	for _, user := range users {
		newusers = append(newusers, &authv1.User{
			Id:    user.ID.String(),
			Email: user.Email,
			Title: user.Title,
			Role:  user.Role,
		})
	}
	return newusers, nil
}

func (s *Service) GetUserById(ctx context.Context, id uuid.UUID) (*authv1.User, error) {

	const op = "Auth.Service.GetUserById"
	log := s.log.With("op", op)

	user, err := s.authrepository.GetUserById(ctx, id)
	if err != nil {
		log.Error("failed to get user by id", "error", err)
	}
	return &authv1.User{
		Id:    user.ID.String(),
		Email: user.Email,
		Title: user.Title,
		Role:  user.Role,
	}, nil
}
