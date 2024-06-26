package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

type closeable interface {
	Close() error
}

type CredentialsStorager interface {
	closeable
	All(ctx context.Context) ([]models.Credentials, error)
	ByLogin(ctx context.Context, login string) (models.Credentials, error)
	Save(ctx context.Context, cred models.Credentials) error
	Update(ctx context.Context, cred models.Credentials) error
}

type TextStorager interface {
	closeable
	All(ctx context.Context) ([]models.Text, error)
	ByKey(ctx context.Context, key string) (models.Text, error)
	Save(ctx context.Context, text models.Text) error
	Update(ctx context.Context, text models.Text) error
}

type BinaryStorager interface {
	closeable
	All(ctx context.Context) ([]models.Binary, error)
	ByKey(ctx context.Context, key string) (models.Binary, error)
	Save(ctx context.Context, bin models.Binary) error
	Update(ctx context.Context, bin models.Binary) error
}

type CardStorager interface {
	closeable
	All(ctx context.Context) ([]models.Card, error)
	ByNumber(ctx context.Context, number string) (models.Card, error)
	Save(ctx context.Context, card models.Card) error
	Update(ctx context.Context, card models.Card) error
}

type Keeper struct {
	log       *slog.Logger
	ch        chan models.Message
	credStore CredentialsStorager
	textStore TextStorager
	binStore  BinaryStorager
	cardStore CardStorager
}

func NewKeeper(log *slog.Logger, ch chan models.Message, credStore CredentialsStorager,
	textStore TextStorager, binStore BinaryStorager, cardStore CardStorager) *Keeper {

	return &Keeper{
		log:       log,
		ch:        ch,
		credStore: credStore,
		textStore: textStore,
		binStore:  binStore,
		cardStore: cardStore,
	}
}

func (s *Keeper) ApplyMessage(ctx context.Context, msg models.Message) {
	switch msg.Type {
	case models.Update:
		s.apply(ctx, msg.Value)
	case models.Snapshot:
		var values [][]byte
		_ = json.Unmarshal(msg.Value, &values)

		for _, value := range values {
			s.apply(ctx, value)
		}
	}
}

func (s *Keeper) Stop() {
	const op = "service.Keeper.Stop"
	log := s.log.With(
		slog.String("op", op),
	)
	if err := s.binStore.Close(); err != nil {
		log.Error("failed to close database connection for binary storage")
	}
	if err := s.cardStore.Close(); err != nil {
		log.Error("failed to close database connection for card storage")
	}
	if err := s.textStore.Close(); err != nil {
		log.Error("failed to close database connection for text storage")
	}
	if err := s.credStore.Close(); err != nil {
		log.Error("failed to close database connection for credentials storage")
	}
}

func (s *Keeper) apply(ctx context.Context, value []byte) {
	const op = "service.Keeper.ApplyMessage"
	log := s.log.With(
		slog.String("op", op),
	)

	var header struct{ Type string }
	_ = json.Unmarshal(value, &header)

	switch header.Type {
	case models.CredItem.String():
		var cred models.Credentials
		_ = json.Unmarshal(value, &cred)
		if err := s.saveCredentials(ctx, cred); err != nil {
			log.Error("apply credentials message error", logger.Err(err))
		}

	case models.TextItem.String():
		var text models.Text
		_ = json.Unmarshal(value, &text)
		if err := s.saveText(ctx, text); err != nil {
			log.Error("apply text message error", logger.Err(err))
		}

	case models.BinItem.String():
		var bin models.Binary
		_ = json.Unmarshal(value, &bin)
		if err := s.saveBinary(ctx, bin); err != nil {
			log.Error("apply binary message error", logger.Err(err))
		}

	case models.CardItem.String():
		var card models.Card
		_ = json.Unmarshal(value, &card)
		if err := s.saveCard(ctx, card); err != nil {
			log.Error("apply card message error", logger.Err(err))
		}
	}
}
