package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func (s *Keeper) SendSaveCard(ctx context.Context, card models.Card) error {
	s.ch <- s.cardToMsg(card)

	return s.saveCard(ctx, card)
}

func (s *Keeper) saveCard(ctx context.Context, card models.Card) error {
	const op = "service.Card.Save"
	log := s.log.With(
		slog.String("op", op),
	)

	if _, err := s.cardStore.ByNumber(ctx, card.Number); err == nil {
		if err = s.cardStore.Update(ctx, card); err != nil {
			log.Error("update card error", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrInternal)
		}

		return nil
	}

	if err := s.cardStore.Save(ctx, card); err != nil {
		log.Error("save card error", logger.Err(err))
		return fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return nil
}

func (s *Keeper) AllCard(ctx context.Context) ([]models.Card, error) {
	const op = "service.Card.All"
	log := s.log.With(
		slog.String("op", op),
	)

	cards, err := s.cardStore.All(ctx)
	if err != nil {
		log.Error("query all cards error", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return cards, nil
}

func (s *Keeper) ValidateCard(card models.Card) ([]string, bool) {
	msg := []string{}
	if card.Number == "" {
		msg = append(msg, "card number should not be empty")
	}
	if err := goluhn.Validate(card.Number); err != nil && card.Number != "" {
		msg = append(msg, "card number should pass digit check (Luhn algorithm)")
	}
	return msg, len(msg) == 0
}

func (s *Keeper) cardToMsg(card models.Card) models.Message {
	value, _ := json.Marshal(card)
	return models.Message{
		Type:  "new",
		Value: value,
	}
}
