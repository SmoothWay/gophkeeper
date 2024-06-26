package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

type TextStorage struct {
	db      *sql.DB
	timeout time.Duration
}

func NewText(storagePath string, timeout time.Duration) (*TextStorage, error) {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return nil, err
	}

	return &TextStorage{
		db:      db,
		timeout: timeout,
	}, nil
}

func (t *TextStorage) All(ctx context.Context) ([]models.Text, error) {
	const op = "storage.Text.All"

	newCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	stmt, err := t.db.Prepare("SELECT tag, key, value, comment, created_at FROM text")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(newCtx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := make([]models.Text, 1)
	for rows.Next() {
		text := models.Text{}
		err = rows.Scan(&text.Tag, &text.Key, &text.Value, &text.Comment)
		if err != nil {
			continue
		}
		res = append(res, text)
	}
	return res, nil
}

func (t *TextStorage) ByKey(ctx context.Context, key string) (models.Text, error) {
	const op = "storage.Text.ByKey"

	newCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	stmt, err := t.db.Prepare("SELECT tag, key, value, comment, created_at FROM text WHERE key = ?")
	if err != nil {
		return models.Text{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, key)
	text := models.Text{Type: models.TextItem}

	err = row.Scan(&text.Tag, &text.Key, &text.Value, &text.Comment, &text.Created)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Text{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}
		return models.Text{}, fmt.Errorf("%s, %w", op, err)
	}

	return text, nil
}

func (t *TextStorage) Save(ctx context.Context, text models.Text) error {
	const op = "storage.Text.Save"

	newCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	stmt, err := t.db.Prepare("INSERT INTO text(tag, key, value, comment, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, text.Tag, text.Key, text.Value, text.Comment, text.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (t *TextStorage) Update(ctx context.Context, text models.Text) error {
	const op = "storage.Text.Update"

	newCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	stmt, err := t.db.Prepare("UPDATE text SET tag=?, value=?, cmment=?, created_at=? WHERE key=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, text.Tag, text.Value, text.Comment, text.Created, text.Key)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (t *TextStorage) Close() error {
	if err := t.db.Close(); err != nil {
		return ErrInternalError
	}
	return nil
}
