package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const VerifyCodeTTL = 5 * time.Minute

var ErrVerifyCodeNotFound = errors.New("verify code not found or expired")

type VerifyRepository struct {
	rdb *redis.Client
	log *slog.Logger
}

func NewVerifyRepository(log *slog.Logger, rdb *redis.Client) *VerifyRepository {
	return &VerifyRepository{rdb: rdb, log: log}
}

func verifyKey(email string) string {
	return "auth:verify:" + email
}

func (r *VerifyRepository) Save(ctx context.Context, email string, code int) error {
	if err := r.rdb.Set(ctx, verifyKey(email), code, VerifyCodeTTL).Err(); err != nil {
		r.log.Error("failed to save verify code", "error", err, "email", email)
		return fmt.Errorf("verify save: %w", err)
	}
	return nil
}

func (r *VerifyRepository) Get(ctx context.Context, email string) (int, error) {
	raw, err := r.rdb.Get(ctx, verifyKey(email)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrVerifyCodeNotFound
	}
	if err != nil {
		r.log.Error("failed to get verify code", "error", err, "email", email)
		return 0, fmt.Errorf("verify get: %w", err)
	}
	code, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("verify parse: %w", err)
	}
	return code, nil
}

func (r *VerifyRepository) Delete(ctx context.Context, email string) error {
	if err := r.rdb.Del(ctx, verifyKey(email)).Err(); err != nil {
		r.log.Error("failed to delete verify code", "error", err, "email", email)
		return fmt.Errorf("verify delete: %w", err)
	}
	return nil
}
