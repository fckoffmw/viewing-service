package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"w2g/internal/auth"

	apperrors "w2g/internal/errors"

	"github.com/google/uuid"
)

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(apperrors.ErrorResponse{Error: msg})
}

type SessionStore interface {
	Get(id string) (*auth.Session, bool)
	Set(session *auth.Session)
}

func Auth(log *slog.Logger, sessionStore SessionStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// Public routes - skip auth
		publicPaths := map[string]bool{
			"/":                  true,
			"/index.html":          true,
			"/login.html":        true,
			"/register.html":     true,
			"/login":           true,
			"/register":        true,
			"/auth/login":     true,
			"/auth/register":   true,
			"/auth/logout":     true,
			"/auth/me":         true,
			"/healthz":         true,
			"/ws":             true,
		}
		if r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/demo/") {
			next.ServeHTTP(w, r)
			return
		}
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		requestID := uuid.New().String()[:uuidLen]

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

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

		if session.ExpiresAt.Before(time.Now()) {
			writeUnauthorized(w, "session expired")
			return
		}

		now := time.Now()
		session.LastSeenAt = now
		session.ExpiresAt = now.Add(auth.SessionExpiry)
		sessionStore.Set(session)

		ctx = context.WithValue(r.Context(), "user_id", session.UserID)
		r = r.WithContext(ctx)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		log.Info("request",
			"user_id", session.UserID,
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}
