package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Keeper struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

func NewKeeperPostgres(db *pgxpool.Pool, timeout time.Duration) *Keeper {
	return &Keeper{
		db:      db,
		timeout: timeout,
	}
}

// Snapshot collect all actual user data with unique keys.
func (s *Keeper) Snapshot(ctx context.Context, userID int64) ([]Item, error) {
	const op = "storage.server.Snapshot"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	rows, err := s.db.Query(newCtx,
		`select (t1.user_id, t1.type, t1.key, s.data, t1.created_at_client) from
		(select user_id, type, key, max(created_at_client) as created_at_client from store where user_id=$1 group by user_id, type, key) as t1
		left join store as s ON t1.type = s.type AND t1.key = s.key AND t1.created_at_client=s.created_at_client`, userID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	res, err := pgx.CollectRows(rows, pgx.RowTo[Item])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return res, nil
}

// Save method insert into database user encrypted message.
func (s *Keeper) Save(ctx context.Context, item Item) error {
	const op = "storage.server.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := s.db.Exec(newCtx, "INSERT INTO store (user_id, type, key, data, created_at_client) values ($1, $2, $3, $4, $5);",
		item.UserID, item.Kind, item.Key, item.Data, item.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
