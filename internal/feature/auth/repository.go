package auth

import (
	"context"
	constants "github.com/Fitnow08/fitnow-backend-sso/internal/models/constants"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type Repository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}
type UserDB struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Email    string    `json:"email" db:"email"`
	Password []byte    `json:"password" db:"password"`
	Title    string    `json:"title" db:"title"`
}

func NewRepository(log *slog.Logger, db *pgxpool.Pool) *Repository {
	return &Repository{log: log, db: db}
}

func (r *Repository) CreateUser(ctx context.Context, email, title string, password []byte) (*UserDB, error) {
	query, arg, err := sq.Insert(constants.UsersTableName).
		Columns("title", "password", "email").
		Values(title, password, email).
		Suffix("RETURNING id, email,title").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		r.log.Error("failed create user", "error", err.Error())
		return nil, err
	}
	var user UserDB

	if err := r.db.QueryRow(ctx, query, arg...).Scan(&user.ID, &user.Email, &user.Title); err != nil {
		r.log.Error("failed create user", "error", err.Error())
		return nil, err
	}
	return &user, nil
}
func (r *Repository) SetVerifyAccount(ctx context.Context, email string) error {
	query, args, err := sq.Update(constants.UsersTableName).
		Set("is_verified", true).
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		r.log.Error("failed build set verify account query", "error", err.Error())
		return err
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		r.log.Error("failed set verify account", "error", err.Error())
		return err
	}
	return nil
}
func (r *Repository) UserByEmail(ctx context.Context, email string) (*UserDB, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	query, args, err := sq.
		Select("id,password,email,title").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		r.log.Warn("failed get user by email", "error", err.Error())
		return nil, err
	}
	var user UserDB
	if err := r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Password, &user.Email, &user.Title); err != nil {
		r.log.Error("failed get user by email sql", "error", err.Error())
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, email string, password []byte) error {
	query, args, err := sq.Update(constants.UsersTableName).
		Set("password", password).
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		r.log.Error("failed update user", "error", err.Error())
		return err
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		r.log.Error("failed update user", "error", err.Error())
		return err
	}
	return nil
}
