package realtime

import "time"

type chatPayload struct {
	Text string `json:"text"`
}

type playerPayload struct {
	Position float64 `json:"position"`
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

func newSyncPayload(s state) syncPayload {
	return syncPayload{
		SourceID:  s.SourceID,
		SourceURL: s.SourceURL,
		Playing:   s.Playing,
		Position:  s.Position,
		UpdatedAt: s.UpdatedAt,
	}
}
