package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/SmoothWay/gophkeeper/internal/server/clients"
	"github.com/SmoothWay/gophkeeper/pkg/logger"
	"github.com/SmoothWay/gophkeeper/pkg/models"
	"github.com/gorilla/websocket"
)

type IService interface {
	Snapshot(ctx context.Context, userID int64) (models.Message, error)
	Save(ctx context.Context, userID int64, msg models.Message) error
	Validate(msg models.Message) (models.Message, error)
}

// Handler handle request for establish connection from user.
// Handler sends and receives user messages.
type Handler struct {
	log        *slog.Logger
	service    IService
	wsUpgrader *websocket.Upgrader
	conns      *clients.UserConnMap
}

func NewHandler(log *slog.Logger, s IService, conns *clients.UserConnMap) *Handler {
	return &Handler{
		log:        log,
		service:    s,
		wsUpgrader: &websocket.Upgrader{},
		conns:      conns,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	op := "ws.Handle"
	log := h.log.With(
		slog.String("op", op),
	)

	ctx := r.Context()

	conn, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(
			"failed establish websocket connection",
			logger.Err(err),
		)
		return
	}

	token := r.Header.Get("token")
	userID, err := lib.ParseToken(token)
	if err != nil {
		log.Error(
			"invalid token",
			slog.String("token", token),
			logger.Err(err),
		)
		errMsg, _ := json.Marshal(models.Message{Type: "error", Value: []byte("invalid token")})
		err = conn.WriteMessage(websocket.TextMessage, errMsg)
		_ = conn.Close()
		return
	}

	h.conns.Put(userID, conn)
	snapshot, err := h.service.Snapshot(ctx, userID)
	if err != nil {
		log.Error(
			"failed collect init snapshot data for user",
			slog.Int64("user_id", userID),
			logger.Err(err),
		)
		errMsg, _ := json.Marshal(models.Message{Type: "error", Value: []byte("failed collect init snapshot data")})
		err = conn.WriteMessage(websocket.TextMessage, errMsg)

		if err != nil {
			// TODO handle interrupted connection with client
			log.Error(
				"error sending message to user",
				slog.Int64("user_id", userID),
				slog.String("address", conn.RemoteAddr().String()),
				logger.Err(err),
			)
		}
	}
	msg, _ := json.Marshal(snapshot)
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		// TODO handle interrupted connection with client
		log.Error(
			"error sending message to user",
			slog.Int64("user_id", userID),
			slog.String("address", conn.RemoteAddr().String()),
			logger.Err(err),
		)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("client logged out")
			// TODO clear user_id - conn map
			return
		default:
			mt, data, err := conn.ReadMessage()
			if err != nil {
				// TODO handle interrupted connection with client
				log.Error(
					"error listening client connection",
					slog.Int64("user_id", userID),
					slog.String("address", conn.RemoteAddr().String()),
					logger.Err(err),
				)
				continue
			}
			if mt != websocket.TextMessage {
				log.Info(
					"unexpected ws message type",
					slog.Int64("user_id", userID),
					slog.Int("websocket message type", mt),
				)
				continue
			}

			var mesg models.Message
			if err := json.Unmarshal(data, &mesg); err != nil {
				log.Info(
					"message cannot be converted into models.Message",
					slog.Int64("user_id", userID),
					slog.String("message", string(data)),
					logger.Err(err),
				)
				continue
			}

			_, err = lib.ParseToken(mesg.Token)
			if err != nil {
				log.Error(
					"invalid token",
					slog.String("token", mesg.Token),
					logger.Err(err),
				)
				errMsg, _ := json.Marshal(models.Message{Type: "error", Value: []byte("invalid token")})
				_ = conn.WriteMessage(websocket.TextMessage, errMsg)
				// TODO clear user_id - conn map
				return
			}

			updateMsg, err := h.service.Validate(mesg)
			if err != nil {
				log.Error(
					"invalid message",
					slog.Int64("user_id", userID),
					slog.String("message", string(mesg.Value)),
					logger.Err(err),
				)
				continue
			}
			err = h.service.Save(ctx, userID, mesg)
			if err != nil {
				log.Error(
					"error saving message into database",
					slog.Int64("user_id", userID),
					slog.String("message", string(mesg.Value)),
					logger.Err(err),
				)
				continue
			}

			go h.sendUpdates(userID, updateMsg)
		}
	}

}

func (h *Handler) sendUpdates(userID int64, msg models.Message) {
	update, _ := json.Marshal(msg)
	for _, c := range h.conns.UserCons(userID) {
		err := c.WriteMessage(websocket.TextMessage, update)
		if err != nil {
			// TODO ? clear user_id - conn map
			h.log.Error(
				"error listening client connection",
				slog.Int64("user_id", userID),
				slog.String("address", c.RemoteAddr().String()),
				logger.Err(err),
			)
			continue
		}
	}
}
