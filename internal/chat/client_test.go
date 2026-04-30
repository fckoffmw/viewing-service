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