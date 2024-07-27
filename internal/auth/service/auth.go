package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/SmoothWay/gophkeeper/internal/auth/storage"
	"github.com/SmoothWay/gophkeeper/pkg/jwt"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidData        = errors.New("invalid request")
)

//go:generate mockgen -source=auth.go -destination=../storage/mocks/mock.go
type UserProvider interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	User(ctx context.Context, email string) (models.User, error)
	Close()
}

type AppProvider interface {
	App(ctx context.Context, id int) (models.App, error)
	Close()
}

// Auth implements Auth interface (grpcapp module).
type Auth struct {
	log          *slog.Logger
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(log *slog.Logger, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Register method hashes password and saves user data into database.
// It returns ErrUserExists, if user with email already registered.
func (a *Auth) Register(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "auth.Register"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	if err := validate(email, password); err != nil {
		return 0, err
	}

	log.Debug("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userProvider.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("user registered")
	return id, nil
}

// Login method checks credentials and returns JWT token.
// It returns ErrInvalidCredentials, if user with credentials does not registered.
func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	if err := validate(email, password); err != nil {
		return "", err
	}

	log.Debug("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", logger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func validate(email string, password string) error {
	switch {
	case email == "":
		return fmt.Errorf("%s, %w", "email is required", ErrInvalidData)
	case password == "":
		return fmt.Errorf("%s, %w", "password is required", ErrInvalidData)
	default:
		return nil
	}
}

func (a *Auth) Close() {
	a.userProvider.Close()
	a.appProvider.Close()
}
