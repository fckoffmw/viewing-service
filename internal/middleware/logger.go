package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"w2g/internal/utils/ctx"

	"github.com/google/uuid"
)

const (
	uuidLen = 8
)

func Logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.New().String()[:uuidLen]

		r = r.WithContext(ctx.WithRequestID(r.Context(), requestID))

		w.Header().Set("X-Request-ID", requestID)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		log.Info("request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}
