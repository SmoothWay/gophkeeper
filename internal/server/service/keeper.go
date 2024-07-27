package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/internal/server/storage"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrMakeSnapshot   = errors.New("get snapshot error")
	ErrInternal       = errors.New("internal error")
)

//go:generate mockgen -source=keeper.go -destination=../storage/mocks/mock.go
type Storager interface {
	Snapshot(ctx context.Context, userID int64) ([]storage.Item, error)
	Save(ctx context.Context, item storage.Item) error
}

type Service struct {
	log     *slog.Logger
	storage Storager
	key     string
}

func New(log *slog.Logger, s Storager, key string) *Service {
	return &Service{
		log:     log,
		storage: s,
		key:     key,
	}
}

func (s *Service) Snapshot(ctx context.Context, userID int64) (models.Message, error) {
	const op = "servicekeeper.Snapshot"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	res, err := s.storage.Snapshot(ctx, userID)
	if err != nil {
		log.Error(
			"query snapshot error",
			logger.Err(err),
		)
		return models.Message{}, fmt.Errorf("%s: %w", op, ErrMakeSnapshot)
	}

	return s.convertItemListToMessage(res), nil
}

func (s *Service) Save(ctx context.Context, userID int64, msg models.Message) error {
	const op = "servicekeeper.Save"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	item := s.convertMessageToItem(userID, msg)
	if err := s.storage.Save(ctx, item); err != nil {
		log.Error(
			"saving new item error",
			slog.String("item type", item.Kind),
			slog.String("item key", item.Key),
			logger.Err(err),
		)
		return ErrInternal
	}

	return nil
}
