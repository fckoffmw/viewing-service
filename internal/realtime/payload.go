package realtime

import "time"

type chatPayload struct {
	Text string `json:"text"`
}

type playerPayload struct {
	Position float64 `json:"position"`
	Playing  *bool   `json:"playing,omitempty"`
}

type syncPayload struct {
	SourceID  string    `json:"source_id"`
	SourceURL string    `json:"source_url"`
	Playing   bool      `json:"playing"`
	Position  float64   `json:"position"`
	UpdatedAt time.Time `json:"updated_at"`
}

type sourceChangedPayload struct {
	SourceID  string `json:"source_id"`
	SourceURL string `json:"source_url"`
}

type stickerPayload struct {
	StickerID string `json:"id"`
}

func newSyncPayload(s state) syncPayload {
	return syncPayload(s)
}
