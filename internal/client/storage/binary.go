package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

type BinaryStorage struct {
	db      *sql.DB
	timeout time.Duration
}

func NewBinary(storagePath string, timeout time.Duration) (*BinaryStorage, error) {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return nil, err
	}

	return &BinaryStorage{
		db:      db,
		timeout: timeout,
	}, nil
}

func (b *BinaryStorage) All(ctx context.Context) ([]models.Binary, error) {
	const op = "storage.Binary.All"

	newCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	stmt, err := b.db.Prepare("SELECT tag, key, value, comment, created_at FROM binary")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(newCtx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	bins := []models.Binary{}

	for rows.Next() {
		bin := models.Binary{Type: models.BinItem}
		err = rows.Scan(&bin.Tag, &bin.Key, &bin.Value, &bin.Comment, &bin.Created)
		if err != nil {
			continue
		}
		bins = append(bins, bin)
	}
	return bins, nil
}

func (b *BinaryStorage) ByKey(ctx context.Context, key string) (models.Binary, error) {
	const op = "storage.Binary.ByKey"

	newCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	stmt, err := b.db.Prepare("SELECT tag, key, value, comment, created_at FROM binary WHERE key = ?")
	if err != nil {
		return models.Binary{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(newCtx, key)

	bin := models.Binary{Type: models.BinItem}
	err = row.Scan(&bin.Tag, &bin.Key, &bin.Value, &bin.Comment, &bin.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Binary{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}

		return models.Binary{}, fmt.Errorf("%s: %w", op, err)
	}
	return bin, nil
}

func (b *BinaryStorage) Save(ctx context.Context, bin models.Binary) error {
	const op = "storage.Binary.Save"

	newCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	stmt, err := b.db.Prepare("INSERT INTO binary(tag, key, value, comment, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(newCtx, bin.Tag, bin.Key, bin.Value, bin.Comment, bin.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *BinaryStorage) Update(ctx context.Context, bin models.Binary) error {
	const op = "storage.Binary.Update"

	newCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	stmt, err := b.db.Prepare("UPDATE binary SET tag = ?, value = ?, comment = ?, created_at = ? WHERE key = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(newCtx, bin.Tag, bin.Value, bin.Comment, bin.Created, bin.Key)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *BinaryStorage) Close() error {
	if err := b.db.Close(); err != nil {
		return ErrInternalError
	}

	return nil
}
