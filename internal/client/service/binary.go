package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

const (
	MB = 1e6
)

var (
	ErrFileNotFound = errors.New("file does not exist")
	ErrFileTooBig   = errors.New("file too big, file size should not exceed 1 MB")
	ErrExtractFile  = errors.New("extract file error")
	ErrInternal     = errors.New("internal error")
)

func (s *Keeper) SendSaveBinary(ctx context.Context, bin models.Binary) error {
	s.ch <- binaryToMsg(bin)

	return s.saveBinary(ctx, bin)
}

func (s *Keeper) saveBinary(ctx context.Context, bin models.Binary) error {
	const op = "service.Binary.Save"
	log := s.log.With(
		slog.String("op", op),
	)

	if _, err := s.binStore.ByKey(ctx, bin.Key); err == nil {
		if err = s.binStore.Update(ctx, bin); err != nil {
			log.Error("update binary error", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrInternal)
		}

		return nil
	}

	if err := s.binStore.Save(ctx, bin); err != nil {
		log.Error("save binary error", logger.Err(err))
		return fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return nil
}

func (s *Keeper) AllBinary(ctx context.Context) ([]models.Binary, error) {
	const op = "service.Binary.All"
	log := s.log.With(
		slog.String("op", op),
	)

	bins, err := s.binStore.All(ctx)
	if err != nil {
		log.Error("query all binary error", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return bins, nil
}

func (s *Keeper) ExtractDataFromFile(path string) (string, []byte, error) {
	const op = "service.Binary.Validate"
	log := s.log.With(
		slog.String("op", op),
		slog.String("file path", path),
	)

	var info os.FileInfo
	var err error
	if info, err = os.Stat(path); os.IsNotExist(err) {
		log.Error("file does not exist")
		return "", nil, fmt.Errorf("%s, %w", op, ErrFileNotFound)
	}

	if info.Size() > MB {
		log.Error(
			"file too big, file size should not exceed 1 MB",
			slog.Int64("file size", info.Size()),
		)
		return "", nil, fmt.Errorf("%s: %w", op, ErrFileTooBig)
	}

	file, err := os.Open(path)
	if err != nil {
		log.Error("open file error", logger.Err(err))
		return "", nil, fmt.Errorf("%s: %w", op, ErrExtractFile)
	}
	defer file.Close()

	buf := []byte{}
	buf, err = io.ReadAll(file)
	if err != nil {
		log.Error("extract file error", logger.Err(err))
		return "", nil, fmt.Errorf("%s: %w", op, ErrExtractFile)
	}

	return info.Name(), buf, nil
}

func ValidateBinary(bin models.Binary) ([]string, bool) {
	msg := []string{}
	if bin.Key == "" {
		msg = append(msg, "file path cannot be empty")
	}
	if len(bin.Value) == 0 {
		msg = append(msg, "file cannot be empty")
	}
	return msg, len(msg) == 0
}

func binaryToMsg(bin models.Binary) models.Message {
	value, _ := json.Marshal(bin)
	return models.Message{
		Type:  "new",
		Value: value,
	}
}
