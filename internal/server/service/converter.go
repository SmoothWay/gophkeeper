package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/pkg/encrypt"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

func (s *Service) Validate(msg models.Message) (models.Message, error) {
	const op = "servicekeeper.Validate"
	log := s.log.With(
		slog.String("op", op),
	)

	var kind struct{ Type string }
	err := json.Unmarshal(msg.Value, &kind)
	if err != nil {
		log.Error(
			"failed extract value from message",
			slog.String(`msg should contains "Type" field`, string(msg.Value)),
			logger.Err(err),
		)
		return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
	}
	switch kind.Type {
	case models.CredItem.String():
		var cred models.Credentials
		err := json.Unmarshal(msg.Value, &cred)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Credentials" model`,
				slog.String(`msg`, string(msg.Value)),
				logger.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case models.TextItem.String():
		var text models.Text
		err = json.Unmarshal(msg.Value, &text)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Text" model`,
				slog.String(`msg`, string(msg.Value)),
				logger.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case models.BinItem.String():
		var bin models.Binary
		err = json.Unmarshal(msg.Value, &bin)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Binary" model`,
				slog.String(`msg`, string(msg.Value)),
				logger.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case models.CardItem.String():
		var card models.Card
		err = json.Unmarshal(msg.Value, &card)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Card" model`,
				slog.String(`msg`, string(msg.Value)),
				logger.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	default:
		log.Error(
			`failed unmarshal message into existing models`,
			slog.String(`unknown message type`, kind.Type),
		)
		return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
	}
	return models.Message{Type: models.Update, Value: msg.Value}, nil
}

func (s *Service) convertItemListToMessage(items []storage.Item) models.Message {
	const op = "servicekeeper.ConvertItemListToMessage"
	log := s.log.With(
		slog.String("op", op),
	)

	values := make([][]byte, 0, len(items))
	for _, item := range items {
		decoded := encrypt.DecodeMsg(string(item.Data), s.key)
		values = append(values, []byte(decoded))
	}
	msg, _ := json.Marshal(values)
	log.Info(
		"items converted",
		slog.Int("number of added items into message", len(items)),
	)

	return models.Message{Type: models.Snapshot, Value: msg}
}

func (s *Service) convertMessageToItem(userID int64, msg models.Message) storage.Item {
	const op = "servicekeeper.ConvertItemListToMessage"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	var item storage.Item
	item.Data = []byte(encrypt.EncodeMsg(msg.Value, s.key))
	item.UserID = userID

	var kind struct{ Type string }
	_ = json.Unmarshal(msg.Value, &kind)

	switch kind.Type {
	case models.CredItem.String():
		var cred models.Credentials
		_ = json.Unmarshal(msg.Value, &cred)

		item.Kind = encrypt.EncodeMsg([]byte(models.CredItem.String()), s.key)
		item.Key = encrypt.EncodeMsg([]byte(cred.Login), s.key)
		item.CreatedAt = cred.Created

	case models.TextItem.String():
		var text models.Text
		_ = json.Unmarshal(msg.Value, &text)

		item.Kind = encrypt.EncodeMsg([]byte(models.TextItem.String()), s.key)
		item.Key = encrypt.EncodeMsg([]byte(text.Key), s.key)
		item.CreatedAt = text.Created

	case models.BinItem.String():
		var bin models.Binary
		_ = json.Unmarshal(msg.Value, &bin)

		item.Kind = encrypt.EncodeMsg([]byte(models.BinItem.String()), s.key)
		item.Key = encrypt.EncodeMsg([]byte(bin.Key), s.key)
		item.CreatedAt = bin.Created

	case models.CardItem.String():
		var card models.Card
		_ = json.Unmarshal(msg.Value, &card)

		item.Kind = encrypt.EncodeMsg([]byte(models.CardItem.String()), s.key)
		item.Key = encrypt.EncodeMsg([]byte(card.Number), s.key)
		item.CreatedAt = card.Created
	}

	log.Info(
		"message converted",
		slog.String("msg value", string(msg.Value)),
	)

	return item
}
