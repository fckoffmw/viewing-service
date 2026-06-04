package realtime

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"w2g/internal/auth"

	"github.com/gorilla/websocket"
)

type mockAuthService struct {
	user *auth.User
}

func (m *mockAuthService) GetUserBySession(_ string) (*auth.User, error) {
	if m.user == nil {
		return nil, errors.New("not found")
	}
	return m.user, nil
}

type mockHubGetter struct {
	hub *hub
}

func (m *mockHubGetter) GetOrCreate(_ string) *hub {
	return m.hub
}

func TestServeWS_NoCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/test-room", nil)
	w := httptest.NewRecorder()

	ServeWS(slog.Default(), &mockHubGetter{}, &mockAuthService{}, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestServeWS_InvalidSession(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/test-room", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookie, Value: "invalid"})
	w := httptest.NewRecorder()

	ServeWS(slog.Default(), &mockHubGetter{}, &mockAuthService{}, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestServeWS_EmptyInviteCode(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/", nil)
	req.AddCookie(&http.Cookie{Name: sessionIDCookie, Value: "valid"})
	w := httptest.NewRecorder()

	ServeWS(slog.Default(), &mockHubGetter{}, &mockAuthService{
		user: &auth.User{ID: "u1", Username: "testuser"},
	}, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestServeWS_WebSocketUpgrade(t *testing.T) {
	h := newHub(slog.Default(), "test-room")
	go h.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/{invite_code}", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(slog.Default(), &mockHubGetter{hub: h}, &mockAuthService{
			user: &auth.User{ID: "u1", Username: "testuser"},
		}, w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws/test-room"

	header := http.Header{}
	header.Add("Cookie", "session_id=valid")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	//nolint:errcheck
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read sync message: %v", err)
	}

	var sync outgoingMessage
	if err := json.Unmarshal(msg, &sync); err != nil {
		t.Fatalf("failed to unmarshal sync: %v", err)
	}
	if sync.Type != MsgTypeSync {
		t.Errorf("expected sync, got %s", sync.Type)
	}

	//nolint:errcheck
	conn.Close()
	h.Close()
}

func TestClientSend(t *testing.T) {
	h := newHub(slog.Default(), "test-room")
	c := &client{
		hub:  h,
		send: make(chan outgoingMessage, 10),
	}

	ch := c.Send()
	if ch == nil {
		t.Fatal("Send() returned nil")
	}

	msg := outgoingMessage{Type: "test", Payload: "data"}
	ch <- msg
	got := <-ch
	if got.Type != "test" {
		t.Errorf("expected test, got %s", got.Type)
	}
}

func TestChatMessageParse(t *testing.T) {
	tests := []struct {
		name      string
		jsonStr   string
		wantValid bool
	}{
		{
			name:      "valid message",
			jsonStr:   `{"text":"hello"}`,
			wantValid: true,
		},
		{
			name:      "valid message with spaces",
			jsonStr:   `{"text":"  hello world  "}`,
			wantValid: true,
		},
		{
			name:      "invalid json",
			jsonStr:   `invalid`,
			wantValid: false,
		},
		{
			name:      "empty text",
			jsonStr:   `{"text":""}`,
			wantValid: false,
		},
		{
			name:      "text too long",
			jsonStr:   `{"text":"` + strings.Repeat("a", maxTextLen+1) + `"}`,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload chatPayload
			err := json.Unmarshal([]byte(tt.jsonStr), &payload)

			if !tt.wantValid {
				if err != nil {
					return
				}
				payload.Text = strings.TrimSpace(payload.Text)
				isValid := payload.Text != "" && len(payload.Text) <= maxTextLen
				if isValid {
					t.Error("expected invalid message")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			payload.Text = strings.TrimSpace(payload.Text)
			isValid := payload.Text != "" && len(payload.Text) <= maxTextLen

			if !isValid {
				t.Error("message should be valid")
			}
		})
	}
}
