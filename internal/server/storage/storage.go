package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInternal = errors.New("internal error")
)

func New(databaseURL string, timeout time.Duration) (*pgxpool.Pool, error) {
	newCtx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	pool, err := pgxpool.New(newCtx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("init database error: %w", ErrInternal)
	}

	if err = migrate(pool, 1); err != nil {
		return nil, fmt.Errorf("migrate database error: %w", ErrInternal)
	}

	return pool, nil
}
