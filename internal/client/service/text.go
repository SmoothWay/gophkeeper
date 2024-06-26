package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func (s *Keeper) SendSaveText(ctx context.Context, text models.Text) error {
	s.ch <- textToMsg(text)

	return s.saveText(ctx, text)
}

func (s *Keeper) saveText(ctx context.Context, text models.Text) error {
	const op = "service.Text.Save"
	log := s.log.With(
		slog.String("op", op),
	)

	if _, err := s.textStore.ByKey(ctx, text.Key); err == nil {
		if err = s.textStore.Update(ctx, text); err != nil {
			log.Error("update text err", sl.Err(err))
			return fmt.Errorf("%s: %w", op, ErrInternal)
		}

		return err
	}

	if err := s.textStore.Save(ctx, text); err != nil {
		log.Error("save text err", sl.Err(err))
		return fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return nil
}

func (s *Keeper) AllText(ctx context.Context) ([]models.Text, error) {
	const op = "service.Text.All"
	log := s.log.With(
		slog.String("op", op),
	)

	list, err := s.textStore.All(ctx)
	if err != nil {
		log.Error("query all text items error", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return list, nil
}

func ValidateText(text models.Text) ([]string, bool) {
	msg := []string{}
	if text.Key == "" {
		msg = append(msg, "key should not be empty")
	}
	if text.Value == "" {
		msg = append(msg, "value should not be empty")
	}
	return msg, len(msg) == 0
}

func textToMsg(text models.Text) models.Message {
	value, _ := json.Marshal(text)
	return models.Message{
		Type:  "new",
		Value: value,
	}
}
