package realtime

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"w2g/internal/auth"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
	maxTextLen     = 1000
	clientBufSize  = 256

	sessionIDCookie = "session_id"
	inviteCodeParam = "invite_code"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

type incomingEvent struct {
	Username string
	Sender   sender
	Message  incomingMessage
}

type client struct {
	hub      *hub
	conn     *websocket.Conn
	send     chan outgoingMessage
	username string
	userID   string
}

type authService interface {
	GetUserBySession(sessionID string) (*auth.User, error)
}

type HubGetter interface {
	GetOrCreate(roomID string) *hub
}

func ServeWS(log *slog.Logger, hubManager HubGetter, authSvc authService, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionIDCookie)
	if err != nil {
		log.Debug("ws: no session cookie")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := authSvc.GetUserBySession(cookie.Value)
	if err != nil || user == nil {
		log.Debug("ws: invalid session")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	inviteCode := r.PathValue(inviteCodeParam)
	if inviteCode == "" {
		log.Debug("ws: missing invite_code in path")
		http.Error(w, "missing invite code", http.StatusBadRequest)

		return
	}

	roomHub := hubManager.GetOrCreate(inviteCode)
	if roomHub == nil {
		log.Error("ws: failed to get or create hub")
		http.Error(w, "internal error", http.StatusInternalServerError)

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("ws: upgrade failed", "err", err)

		return
	}

	client := &client{
		hub:      roomHub,
		conn:     conn,
		username: user.Username,
		userID:   user.ID,
		send:     make(chan outgoingMessage, clientBufSize),
	}

	log.Info("ws client connected", "username", user.Username, "user_id", user.ID, "room_id", inviteCode)

	roomHub.Register() <- client

	go client.writePump(log)
	go client.readPump(log, roomHub)
}

func (c *client) Send() chan outgoingMessage {
	return c.send
}

//nolint:errcheck
func (c *client) readPump(log *slog.Logger, roomHub *hub) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("ws: readPump panic", "recover", r, "user_id", c.userID)
		}
		roomHub.Unregister() <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, reader, err := c.conn.NextReader()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug("ws: unexpected close", "err", err)
			}
			break
		}

		var msg incomingMessage
		if err := json.NewDecoder(reader).Decode(&msg); err != nil {
			log.Debug("ws: decode error", "err", err)
			continue
		}

		log.Info("ws message received",
			"type", msg.Type,
			"username", c.username,
			"payload", string(msg.Payload),
		)

		evt := incomingEvent{
			Username: c.username,
			Sender:   c,
			Message:  msg,
		}

		roomHub.Incoming() <- evt
	}
}

//nolint:errcheck
func (c *client) writePump(log *slog.Logger) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if r := recover(); r != nil {
			log.Error("ws: writePump panic", "recover", r, "user_id", c.userID)
		}
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Debug("ws: write error", "err", err)
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Error("ws: marshal error", "err", err)

				return
			}
			w.Write(data)

			if err := w.Close(); err != nil {
				log.Debug("ws: write close error", "err", err)

				return
			}

			log.Debug("ws message sent", "type", msg.Type, "username", msg.Username, "payload", string(data))
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
