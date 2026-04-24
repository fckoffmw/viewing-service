package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

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

type BroadcasterHub interface {
	Broadcast() chan []byte
}

type Server struct {
	Log *slog.Logger
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	Send chan []byte
}

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type chatMessage struct {
	ClientID string `json:"clientId,omitempty"`
	Text     string `json:"text"`
}

type syncEvent struct {
	Type string  `json:"type"`
	Time float64 `json:"time"`
}

func NewServer(log *slog.Logger) *Server {
	return &Server{
		Log: log,
	}
}

func (s *Server) ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Log.Error("ws upgrade", "err", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		Send: make(chan []byte, 64),
	}

	accepted := make(chan bool, 1)
	hub.register <- registerRequest{client: client, accepted: accepted}
	if !<-accepted {
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "room is full"),
			time.Now().Add(writeWait),
		)
		_ = conn.Close()
		return
	}

	s.Log.Debug("ws client connected", "clients", len(hub.clients))
	go client.writePump(s.Log)
	go client.readPump(s)
}

func (c *Client) readPump(s *Server) {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
		s.Log.Debug("ws client disconnected", "clients", len(c.hub.clients))
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

		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		switch event.Type {
		case "chat_message":
			var payload chatMessage
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				continue
			}
			payload.Text = strings.TrimSpace(payload.Text)
			if payload.Text == "" || len(payload.Text) > maxTextLen {
				continue
			}
			normalized, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			s.Log.Debug("ws chat message", "clientId", payload.ClientID, "text", payload.Text)
			c.hub.broadcast <- normalized

		case "video_sync":
			var payload syncEvent
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				continue
			}
			s.Log.Debug("ws sync event", "type", payload.Type, "time", payload.Time)
			c.hub.broadcast <- message
		}
	}
}

func (c *Client) writePump(log *slog.Logger) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
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
