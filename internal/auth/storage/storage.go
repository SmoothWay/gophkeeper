package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

func Migrate(databaseURL string, timeout time.Duration) error {
	pool, err := newPool(databaseURL, timeout)
	if err != nil {
		return err
	}

	if err = migrate(pool, 1); err != nil {
		return fmt.Errorf("postgres migration error: %w", err)
	}

	pool.Close()

	return nil
}

func newPool(databaseURL string, timeout time.Duration) (*pgxpool.Pool, error) {
	newCtx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	pool, err := pgxpool.New(newCtx, databaseURL)
	if err != nil {
		return pool, fmt.Errorf("postgres connection: %w", err)
	}
	return pool, nil
}
