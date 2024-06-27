package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

type Credentials struct {
	db      *sql.DB
	timeout time.Duration
}

func NewCredentials(storagePath string, timeout time.Duration) (*Credentials, error) {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return nil, err
	}
	return &Credentials{
		db:      db,
		timeout: timeout,
	}, nil
}

func (s *Credentials) All(ctx context.Context) ([]models.Credentials, error) {
	const op = "storage.Credentials.All"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, login, password, comment, created_at FROM credentials")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(newCtx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := []models.Credentials{}

	for rows.Next() {
		cred := models.Credentials{Type: models.CredItem}
		err = rows.Scan(&cred.Tag, &cred.Login, &cred.Password, &cred.Comment, &cred.Comment)
		if err != nil {
			continue
		}
		res = append(res, cred)
	}

	return res, nil
}

func (s *Credentials) ByLogin(ctx context.Context, login string) (models.Credentials, error) {
	const op = "storage.sqlite.Credentials.ByLogin"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, login, password, comment, created_at FROM credentials WHERE login = ?")
	if err != nil {
		return models.Credentials{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, login)

	cred := models.Credentials{Type: models.CredItem}
	err = row.Scan(&cred.Tag, &cred.Login, &cred.Password, &cred.Comment, &cred.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Credentials{}, fmt.Errorf("%s, %w", op, ErrItemNotFound)
		}

		return models.Credentials{}, fmt.Errorf("%s: %w", op, err)
	}

	return cred, nil
}

func (s *Credentials) Save(ctx context.Context, cred models.Credentials) error {
	const op = "storage.Credentials.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("INSERT INTO credentials(tag, login, password, comment, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(newCtx, cred.Tag, cred.Login, cred.Password, cred.Comment, cred.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Credentials) Update(ctx context.Context, cred models.Credentials) error {
	const op = "storage.Credentials.Update"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("UPDATE credentials SET tag=?, password=?, comment=?, created_at=? WHERE login=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(newCtx, cred.Tag, cred.Password, cred.Comment, cred.Created, cred.Login)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Credentials) Close() error {
	if err := s.db.Close(); err != nil {
		return ErrInternalError
	}

	return nil
}
