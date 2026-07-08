package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"w2g/internal/apperrors"
	"w2g/internal/auth"
	"w2g/internal/utils/ctx"
)

type Middleware func(http.Handler, ...any) http.Handler

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(apperrors.ErrorResponse{Error: msg})
}

type SessionStore interface {
	Get(id string) (*auth.Session, bool)
	Set(session *auth.Session)
}

func Auth(log *slog.Logger, next http.Handler, sessionStore SessionStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicPaths := map[string]bool{
			"/":              true,
			"/index.html":    true,
			"/login.html":    true,
			"/register.html": true,
			"/login":         true,
			"/register":      true,
			"/auth/login":    true,
			"/auth/register": true,
			"/auth/logout":   true,
			"/auth/me":       true,
			"/healthz":       true,
		}
		if r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/demo/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			next.ServeHTTP(w, r)

			return
		}
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)

			return
		}

		cookie, err := r.Cookie("session_id")
		if err != nil {
			log.Warn("unauthorized", "reason", "no session cookie")
			writeUnauthorized(w, "session not found")

			return
		}

		session, ok := sessionStore.Get(cookie.Value)
		if !ok {
			log.Warn("unauthorized", "reason", "invalid session")
			writeUnauthorized(w, "session not found")

			return
		}

		now := time.Now()
		session.LastSeenAt = now
		session.ExpiresAt = now.Add(auth.SessionExpiry)
		sessionStore.Set(session)

		r = r.WithContext(ctx.WithUserID(r.Context(), session.UserID))

		next.ServeHTTP(w, r)
	})
}
