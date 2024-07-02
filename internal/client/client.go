package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	viewaddbinary "github.com/SmoothWay/gophkeeper/internal/client/cli/view_add_binary"
	viewaddcard "github.com/SmoothWay/gophkeeper/internal/client/cli/view_add_card"
	viewaddtext "github.com/SmoothWay/gophkeeper/internal/client/cli/view_add_text"
	viewauth "github.com/SmoothWay/gophkeeper/internal/client/cli/view_auth"
	view_command_list "github.com/SmoothWay/gophkeeper/internal/client/cli/view_command_list"
	viewlist "github.com/SmoothWay/gophkeeper/internal/client/cli/view_list"
	viewlogin "github.com/SmoothWay/gophkeeper/internal/client/cli/view_login"
	viewregister "github.com/SmoothWay/gophkeeper/internal/client/cli/view_register"
	"github.com/SmoothWay/gophkeeper/internal/client/config"
	"github.com/SmoothWay/gophkeeper/internal/client/grpcclient"
	"github.com/SmoothWay/gophkeeper/internal/client/service"
	"github.com/SmoothWay/gophkeeper/internal/client/storage"
	"github.com/SmoothWay/gophkeeper/internal/client/ws"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
)

var (
	ErrViewModel      = errors.New("viewing UI model error")
	ErrRetrieveModel  = errors.New("failed retrieve model")
	ErrUserStoppedApp = errors.New("user stopped execution")
)

type AppClient struct {
	ch           chan models.Message
	grpcClient   *grpcclient.GRPCClient
	keeper       *service.Keeper
	log          *slog.Logger
	queryTimeout time.Duration
	storagePath  string
	grpcAddress  string
	WSURL        string
}

func NewAppClient(log *slog.Logger, cfg *config.ClientConfig) *AppClient {
	return &AppClient{
		log:          log,
		storagePath:  cfg.StoragePath,
		grpcAddress:  cfg.GRPCAddress,
		WSURL:        cfg.WSURL,
		queryTimeout: cfg.QueryTime,
	}
}

func (app *AppClient) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"

	log := app.log.With(
		slog.String("op", op),
	)

	app.ch = make(chan models.Message)

	err := storage.Migrate(app.storagePath)
	if err != nil {
		log.Error(
			"migration database error",
			logger.Err(err),
		)
		stop <- syscall.SIGTERM
		return
	}

	dbCred, err := storage.NewCredentials(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init credentials storage")
		stop <- syscall.SIGTERM
		return
	}

	dbText, err := storage.NewText(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init text storage")
		stop <- syscall.SIGTERM
		return
	}

	dbBin, err := storage.NewBinary(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init binary storage")
		stop <- syscall.SIGTERM
		return
	}

	dbCard, err := storage.NewCard(app.storagePath, app.queryTimeout)
	if err != nil {
		log.Error("failed to init card storage")
		stop <- syscall.SIGTERM
		return
	}
	app.keeper = service.NewKeeper(log, app.ch, dbCred, dbText, dbBin, dbCard)

	app.grpcClient, err = grpcclient.NewGRPCClient(app.grpcAddress)
	if err != nil {
		log.Error(
			"failed connect to GRPC auth server",
			slog.String("GRPC address", app.grpcAddress),
			logger.Err(err),
		)
		stop <- syscall.SIGTERM
		return
	}
	p := tea.NewProgram(viewauth.Model{})
	m, _ := p.Run()

	modelAuth, _ := m.(viewauth.Model)
	if modelAuth.Choice == "" {
		log.Info("user stopped execution (q, ctrl+c, esc)")
		stop <- syscall.SIGTERM
		return
	}

	if modelAuth.Choice == "Register" {
		if err := app.registration(ctx); err != nil {
			log.Error("registration failed", logger.Err(err))
			stop <- syscall.SIGTERM
			return
		}
	}

	token, err := app.login(ctx)
	if err != nil || token == "" {
		log.Error("login user error", logger.Err(err))
		stop <- syscall.SIGTERM
		return
	}

	wsClient := ws.NewWSClient(log, app.ch, app.keeper, app.WSURL)

	interrupt := make(chan struct{})
	go func(interrupt chan struct{}) {
		<-interrupt
		stop <- syscall.SIGTERM
	}(interrupt)

	wsClient.Run(ctx, interrupt, token)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			p := tea.NewProgram(view_command_list.Model{})
			m, err := p.Run()
			if err != nil {
				log.Error("viewing command list error", logger.Err(err))
				stop <- syscall.SIGTERM
				return
			}

			modelComandList, ok := m.(view_command_list.Model)
			if !ok {
				log.Error("failed retrieve command list model", logger.Err(err))
				stop <- syscall.SIGTERM
				return
			}

			if modelComandList.Choice == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				log.Info("user stopped execution (q, ctrl+C, esc)")
				stop <- syscall.SIGTERM
				return
			}

			switch modelComandList.Choice {
			case "Get all secrets":
				if err := app.commandGetAllSecrets(ctx); err != nil {
					log.Error("failed execute get all secrets command", logger.Err(err))
					stop <- syscall.SIGTERM
					return
				}

			case "Add credentials":
				ok := app.commandAdd(ctx, app.commandAddCard, "credentials")
				if !ok {
					stop <- syscall.SIGTERM
					return
				}

			case "Add text data":
				ok := app.commandAdd(ctx, app.commandAddText, "text data")
				if !ok {
					stop <- syscall.SIGTERM
					return
				}

			case "Add binary data":
				ok := app.commandAdd(ctx, app.commandAddBinary, "binary data")
				if !ok {
					stop <- syscall.SIGTERM
					return
				}

			case "Add card data":
				ok := app.commandAdd(ctx, app.commandAddCard, "card")
				if !ok {
					stop <- syscall.SIGTERM
					return
				}
			}
		}
	}
}

func (app *AppClient) Stop() {
	app.keeper.Stop()
	close(app.ch)
	app.grpcClient.Stop()
}

func (app *AppClient) registration(ctx context.Context) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			p := tea.NewProgram(viewregister.InitialModel(app.grpcClient))
			m, err := p.Run()
			if err != nil {
				return ErrViewModel
			}

			modelRegister, ok := m.(viewregister.Model)
			if !ok {
				return ErrRetrieveModel
			}

			if modelRegister.State == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				return ErrUserStoppedApp
			}

			if modelRegister.State == "again" {
				continue
			}

			if modelRegister.State == "error" {
				return errors.New("registration failed")
			}

			break loop
		}
	}
	return nil
}

func (app *AppClient) login(ctx context.Context) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", nil
		default:
			p := tea.NewProgram(viewlogin.InitialModel(app.grpcClient))
			m, err := p.Run()
			if err != nil {
				return "", ErrViewModel
			}

			modelLogin, ok := m.(viewlogin.Model)
			if !ok {
				return "", ErrRetrieveModel
			}

			if modelLogin.State == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				return "", ErrUserStoppedApp
			}

			if modelLogin.State == "again" {
				continue
			}

			if modelLogin.State == "error" {
				return "", errors.New("failed login user")
			}

			return modelLogin.Token, nil
		}
	}
}

func (app *AppClient) commandGetAllSecrets(ctx context.Context) error {
	const op = "client.Run.GetAllSecrets"
	log := app.log.With(
		slog.String("op", op),
	)

	var err error
	creds, err := app.keeper.AllCredentials(ctx)
	if err != nil {
		log.Error("query all credentials error", logger.Err(err))
	}
	texts, err := app.keeper.AllText(ctx)
	if err != nil {
		log.Error("query all text data error", logger.Err(err))
	}
	bins, err := app.keeper.AllBinary(ctx)
	if err != nil {
		log.Error("query all binary data error", logger.Err(err))
	}
	cards, err := app.keeper.AllCard(ctx)
	if err != nil {
		log.Error("query all cards error", logger.Err(err))
	}
	// view result
	p := tea.NewProgram(viewlist.Model{Msg: viewlist.Convert(creds, texts, bins, cards)})
	_, err = p.Run()
	if err != nil {
		return ErrViewModel
	}

	return nil
}

func (app *AppClient) commandAdd(ctx context.Context, command func(ctx context.Context) error, msg string) bool {
	const op = "client.Run"
	log := app.log.With(
		slog.String("op", op),
		slog.String("data type", msg),
	)

	if err := command(ctx); err != nil {
		switch {
		case errors.Is(err, ErrUserStoppedApp):
			log.Info("user stopped execution (q, ctrl+C, esc)")
			return false
		case errors.Is(err, ErrViewModel) || errors.Is(err, ErrRetrieveModel):
			log.Error(fmt.Sprintf("add %s error", msg), logger.Err(err))
			return false
		default:
			log.Error(fmt.Sprintf("saving %s error", msg), logger.Err(err))
			return true
		}
	}
	return true
}

func (app *AppClient) commandAddText(ctx context.Context) error {
	p := tea.NewProgram(viewaddtext.InitialModel())
	m, err := p.Run()
	if err != nil {
		return ErrViewModel
	}

	modelAddText, ok := m.(viewaddtext.Model)
	if !ok {
		return ErrRetrieveModel
	}

	if modelAddText.State == "quit" {
		return ErrUserStoppedApp
	}

	text := models.Text{
		Type:    models.TextItem,
		Tag:     modelAddText.Inputs[0].Value(),
		Key:     modelAddText.Inputs[1].Value(),
		Value:   modelAddText.Inputs[2].Value(),
		Comment: modelAddText.Inputs[3].Value(),
		Created: time.Now().Unix(),
	}
	// TODO validate item
	err = app.keeper.SendSaveText(ctx, text)
	if err != nil {
		// TODO view result
		return fmt.Errorf("saving text error %w", err)
	}

	return nil
}

func (app *AppClient) commandAddBinary(ctx context.Context) error {
	p := tea.NewProgram(viewaddbinary.InitialModel())
	m, err := p.Run()
	if err != nil {
		return ErrViewModel
	}

	modelAddBinary, ok := m.(viewaddbinary.Model)
	if !ok {
		return ErrRetrieveModel
	}

	if modelAddBinary.State == "quit" {
		return ErrUserStoppedApp
	}

	path := modelAddBinary.Inputs[1].Value()
	fileName, data, err := app.keeper.ExtractDataFromFile(path)
	if err != nil {
		return fmt.Errorf("failed extract data from file %w", err)
	}

	bin := models.Binary{
		Type:    models.BinItem,
		Tag:     modelAddBinary.Inputs[0].Value(),
		Key:     fileName,
		Value:   data,
		Comment: modelAddBinary.Inputs[2].Value(),
		Created: time.Now().Unix(),
	}
	err = app.keeper.SendSaveBinary(ctx, bin)
	if err != nil {
		return fmt.Errorf("saving binary data error %w", err)
	}

	return nil
}

func (app *AppClient) commandAddCard(ctx context.Context) error {
	p := tea.NewProgram(viewaddcard.InitialModel())
	m, err := p.Run()
	if err != nil {
		return ErrViewModel
	}

	modelAddCard, ok := m.(viewaddcard.Model)
	if !ok {
		return ErrRetrieveModel
	}

	if modelAddCard.State == "quit" {
		// user stopped execution in UI (q, ctrl+C, esc)
		return ErrUserStoppedApp
	}

	cvv, _ := strconv.Atoi(modelAddCard.Inputs[3].Value())
	card := models.Card{
		Type:    models.CardItem,
		Tag:     modelAddCard.Inputs[0].Value(),
		Number:  modelAddCard.Inputs[1].Value(),
		Exp:     modelAddCard.Inputs[2].Value(),
		Cvv:     int32(cvv),
		Comment: modelAddCard.Inputs[4].Value(),
		Created: time.Now().Unix(),
	}
	// TODO validate card item
	err = app.keeper.SendSaveCard(ctx, card)
	if err != nil {
		// TODO view result
		return fmt.Errorf("saving card data error %w", err)
	}

	return nil
}
