package chat

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChatMessageParse(t *testing.T) {
	tests := []struct {
		name      string
		jsonStr   string
		wantID    string
		wantText  string
		wantValid bool
	}{
		{
			name:      "valid message",
			jsonStr:   `{"clientId":"user1","text":"hello"}`,
			wantID:    "user1",
			wantText:  "hello",
			wantValid: true,
		},
		{
			name:      "valid message with spaces",
			jsonStr:   `{"clientId":"user1","text":"  hello world  "}`,
			wantID:    "user1",
			wantText:  "hello world",
			wantValid: true,
		},
		{
			name:      "invalid json",
			jsonStr:   `invalid`,
			wantID:    "",
			wantText:  "",
			wantValid: false,
		},
		{
			name:      "empty text",
			jsonStr:   `{"clientId":"user1","text":""}`,
			wantID:    "user1",
			wantText:  "",
			wantValid: false,
		},
		{
			name:      "text too long",
			jsonStr:   `{"clientId":"user1","text":"` + strings.Repeat("a", maxTextLen+1) + `"}`,
			wantID:    "user1",
			wantText:  "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload chatMessage
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
			if payload.ClientID != tt.wantID {
				t.Errorf("ClientID = %q, want %q", payload.ClientID, tt.wantID)
			}
			if payload.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", payload.Text, tt.wantText)
			}
		})
	}
}

func TestMaxTextLen(t *testing.T) {
	if maxTextLen != 500 {
		t.Errorf("maxTextLen = %d, want 500", maxTextLen)
	}
}

func TestMaxMessageSize(t *testing.T) {
	if maxMessageSize != 1024 {
		t.Errorf("maxMessageSize = %d, want 1024", maxMessageSize)
	}
}
