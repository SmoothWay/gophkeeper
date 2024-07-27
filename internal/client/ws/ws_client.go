package ws

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
	"github.com/gorilla/websocket"
)

var (
	ErrConnectToServer = errors.New("failed establish websocket connection")
)

type MessageService interface {
	ApplyMessage(ctx context.Context, msg models.Message)
}

type WSClient struct {
	log  *slog.Logger
	conn *websocket.Conn
	ch   chan models.Message
	s    MessageService
	url  string
}

func NewWSClient(log *slog.Logger, ch chan models.Message, s MessageService, url string) *WSClient {
	return &WSClient{
		log: log,
		ch:  ch,
		s:   s,
		url: url,
	}
}

func (ws *WSClient) Run(ctx context.Context, interrupt chan struct{}, token string) {
	const op = "ws.Run"
	log := ws.log.With(
		slog.String("op", op),
	)

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	headers := make(map[string][]string)
	headers["token"] = append(headers["token"], token)

	var err error
	ws.conn, _, err = dialer.DialContext(ctx, ws.url, headers)

	if err != nil {
		log.Error(
			"failed establish websocket connection",
			logger.Err(err),
		)
		close(interrupt)
		return
	}

	go ws.read(ctx, interrupt)
	go ws.write(ctx, token, interrupt)
}

func (ws *WSClient) read(ctx context.Context, interrupt chan struct{}) {
	op := "ws.Run.read"
	log := ws.log.With(
		slog.String("op", op),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info(
				"receive context done message",
			)
			return

		default:
			var header struct{ Type string }
			mt, data, err := ws.conn.ReadMessage()
			if err != nil {
				// TODO implement restoring connection to the server
				log.Error(
					"error receiving message from server",
					slog.String("address", ws.conn.RemoteAddr().String()),
					logger.Err(err),
				)
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			err = json.Unmarshal(data, &header)
			if err != nil || (header.Type != "update" && header.Type != "snapshot" && header.Type != "error") {
				continue
			}
			var msg models.Message
			err = json.Unmarshal(data, &msg)
			if err != nil {
				log.Warn(
					"receiving unexpected message from server",
					slog.String("message value", string(msg.Value)),
					logger.Err(err),
				)
				continue
			}
			if msg.Type == "error" && string(msg.Value) == "invalid token" {
				close(interrupt)
				return
			}

			ws.s.ApplyMessage(ctx, msg)
		}
	}
}

func (ws *WSClient) write(ctx context.Context, token string, interrupt chan struct{}) {
	op := "ws.Run.write"
	log := ws.log.With(
		slog.String("op", op),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("recieve context done message")

			return
		case msg := <-ws.ch:
			msg.Token = token
			data, _ := json.Marshal(msg)
			err := ws.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				// TODO implement restoring connection to the server
				close(interrupt)
				return
			}
		}
	}
}
