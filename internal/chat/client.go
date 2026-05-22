package chat

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"w2g/internal/auth"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
	maxTextLen     = 500
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

type chatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

type Client struct {
	hub      *hub
	conn     *websocket.Conn
	send     chan []byte
	username string
	userID   string
}

type AuthService interface {
	GetUserBySession(sessionID string) (*auth.User, error)
}

type HubGetter interface {
	GetOrCreate(roomID string) *hub
}

func ServeWS(log *slog.Logger, hubManager HubGetter, authSvc AuthService, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
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

	inviteCode := r.PathValue("invite_code")
	if inviteCode == "" {
		inviteCode = extractInviteCodeFromPath(r.URL.Path)
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

	client := &Client{
		hub:      roomHub,
		conn:     conn,
		username: user.Username,
		userID:   user.ID,
		send:     make(chan []byte, 256),
	}

	roomHub.Register() <- client

	go client.writePump(log)
	go client.readPump(log, roomHub)
}

func (c *Client) Send() chan []byte {
	return c.send
}

func (c *Client) readPump(log *slog.Logger, roomHub *hub) {
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
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, reader, err := c.conn.NextReader()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug("ws: unexpected close", "err", err)
			}
			break
		}

		var msg chatMessage
		if err := json.NewDecoder(reader).Decode(&msg); err != nil {
			log.Debug("ws: decode error", "err", err)
			continue
		}

		msg.Username = c.username
		msg.Text = strings.TrimSpace(msg.Text)

		if msg.Text == "" {
			continue
		}

		if len(msg.Text) > maxTextLen {
			msg.Text = msg.Text[:maxTextLen]
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Error("ws: marshal error", "err", err)
			continue
		}

		roomHub.Broadcast() <- data
	}
}

func (c *Client) writePump(log *slog.Logger) {
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
			w.Write(msg)

			if err := w.Close(); err != nil {
				log.Debug("ws: write close error", "err", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func extractInviteCodeFromPath(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 && idx+1 < len(path) {
		return path[idx+1:]
	}

	return ""
}
