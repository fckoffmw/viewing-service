package chat

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"w2g/internal/auth"
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

type Client struct {
	hub      *RoomHub
	conn     *websocket.Conn
	send     chan []byte
	username string
	userID   string
}

type chatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

type AuthService interface {
	GetUserBySession(sessionID string) (*auth.User, error)
}

type HubManagerGetter interface {
	GetOrCreate(roomID string) *RoomHub
}

func ServeWS(log *slog.Logger, hubManager HubManagerGetter, authSvc AuthService, w http.ResponseWriter, r *http.Request) {
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

	inviteCode := extractInviteCode(r)
	if inviteCode == "" {
		http.Error(w, "invite code required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade", "err", err)
		return
	}

	roomHub := hubManager.GetOrCreate(inviteCode)

	client := &Client{
		hub:      roomHub,
		conn:     conn,
		send:     make(chan []byte, 64),
		username: user.Username,
		userID:   user.ID,
	}

	roomHub.Register() <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister() <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var payload chatMessage
		if err := json.Unmarshal(message, &payload); err != nil {
			continue
		}

		payload.Text = strings.TrimSpace(payload.Text)
		if payload.Text == "" || len(payload.Text) > maxTextLen {
			continue
		}

		msg := chatMessage{
			Username: c.username,
			Text:     payload.Text,
		}
		normalized, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		c.hub.Broadcast() <- normalized
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func extractInviteCode(r *http.Request) string {
	path := r.URL.Path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return ""
}