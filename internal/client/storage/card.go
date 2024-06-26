package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

type CardStorage struct {
	db      *sql.DB
	timeout time.Duration
}

func NewCard(storagepath string, timeout time.Duration) (*CardStorage, error) {
	db, err := newSQLDB(storagepath)
	if err != nil {
		return nil, err
	}

	return &CardStorage{
		db:      db,
		timeout: timeout,
	}, nil
}

func (c *CardStorage) All(ctx context.Context) ([]models.Card, error) {
	const op = "storage.Card.All"

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stmt, err := c.db.Prepare("SELECT tag, number, exp, cvv, comment, created_at FROM card")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(newCtx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	cards := []models.Card{}

	for rows.Next() {
		card := models.Card{}
		err = rows.Scan(&card.Tag, &card.Number, &card.Exp, &card)
		if err != nil {
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (c *CardStorage) ByNumber(ctx context.Context, number string) (models.Card, error) {
	const op = "storage.Card.ByNumber"

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stmt, err := c.db.Prepare("SELECT tag, number, exp, cvv, comment, created_at FROM card WHERE number = ?")
	if err != nil {
		return models.Card{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, number)

	card := models.Card{Type: models.CardItem}

	err = row.Scan(&card.Tag, &card.Number, &card.Exp, &card.Cvv, &card.Comment, &card.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Card{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}

		return models.Card{}, fmt.Errorf("%s: %w", op, err)
	}

	return card, nil

}

func (c *CardStorage) Save(ctx context.Context, card models.Card) error {
	const op = "storage.Card.Save"

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stmt, err := c.db.Prepare("INSERT INTO card(tag, number, exp, cvv, comment, created_at) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, card.Tag, card.Number, card.Exp, card.Cvv, card.Comment, card.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *CardStorage) Update(ctx context.Context, card models.Card) error {
	const op = "storage.Card.Update"

	newCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stmt, err := c.db.Prepare("UPDATE card SET tag=?, exp=?, cvv=?, comment=?, created_at=? WHERE number=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, card.Tag, card.Exp, card.Cvv, card.Comment, card.Created, card.Number)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CardStorage) Close() error {
	if err := c.db.Close(); err != nil {
		return ErrInternalError
	}

	return nil
}
