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
	//nolint:errcheck
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
			writeUnauthorized(w, "session not found")
			return
		}

		session, ok := sessionStore.Get(cookie.Value)
		if !ok {
			writeUnauthorized(w, "session not found")
			return
		}

		now := time.Now()
		session.LastSeenAt = now
		session.ExpiresAt = now.Add(auth.SessionExpiry)
		sessionStore.Set(session)

		requestID := ctx.RequestIDFromContext(r.Context())
		r = r.WithContext(ctx.WithUserID(r.Context(), session.UserID))

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		log.Info("request",
			"user_id", session.UserID,
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
		)
	})
}
