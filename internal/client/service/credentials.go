package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func (s *Keeper) SendSaveCredentials(ctx context.Context, cred models.Credentials) error {
	s.ch <- credentialsToMsg(cred)

	return s.saveCredentials(ctx, cred)
}

func (s *Keeper) saveCredentials(ctx context.Context, cred models.Credentials) error {
	const op = "service.Credential.Save"
	log := s.log.With(
		slog.String("op", op),
	)

	if _, err := s.credStore.ByLogin(ctx, cred.Login); err == nil {
		if err = s.credStore.Update(ctx, cred); err != nil {
			log.Error("update credentials error", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrInternal)
		}

		return nil
	}

	if err := s.credStore.Save(ctx, cred); err != nil {
		log.Error("save credentials error", logger.Err(err))
		return fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return nil
}

func (s *Keeper) AllCredentials(ctx context.Context) ([]models.Credentials, error) {
	const op = "service.Credential.Save"
	log := s.log.With(
		slog.String("op", op),
	)

	creds, err := s.credStore.All(ctx)
	if err != nil {
		log.Error("query all credentials error", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return creds, nil
}

func ValidateCredentials(cred models.Credentials) ([]string, bool) {
	msg := []string{}
	if cred.Login == "" {
		msg = append(msg, "login should not be empty")
	}
	if cred.Password == "" {
		msg = append(msg, "password should not be empty")
	}
	return msg, len(msg) == 0
}

func credentialsToMsg(cred models.Credentials) models.Message {
	value, _ := json.Marshal(cred)
	return models.Message{
		Type:  "new",
		Value: value,
	}
}
