package http

import (
	"net/http"
	"strings"
)

func ExtractInviteCode(r *http.Request) string {
	if inviteCode := r.PathValue("invite_code"); inviteCode != "" {
		return inviteCode
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rooms/"), "/")
	if len(parts) >= 1 && parts[0] != "" {
		return strings.Split(parts[0], "/")[0]
	}
	return ""
}