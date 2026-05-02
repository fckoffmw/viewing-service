package http

import (
	"log/slog"
	"net/http"

	"w2g/internal/auth"
	"w2g/internal/chat"
	"w2g/internal/middleware"
)

type authHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
}

type sourceHandler interface {
	GetAllSources(w http.ResponseWriter, r *http.Request)
	AddSource(w http.ResponseWriter, r *http.Request)
}

type roomHandler interface {
	CreateRoom(w http.ResponseWriter, r *http.Request)
	GetRoom(w http.ResponseWriter, r *http.Request)
	DeleteRoom(w http.ResponseWriter, r *http.Request)
	PatchRoomSource(w http.ResponseWriter, r *http.Request)
}

type router struct {
	log     *slog.Logger
	Handler http.Handler
}

func NewRouter(
	log *slog.Logger,
	hub *chat.Hub,
	authService auth.Service,
	authHandler authHandler,
	sourceHandler sourceHandler,
	roomHandler roomHandler,
	hubManager *chat.HubManager,
) *router {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws/{invite_code}", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWS(log, hubManager, authService, w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /auth/me", authHandler.Me)

	mux.HandleFunc("GET /api/sources", sourceHandler.GetAllSources)
	mux.HandleFunc("POST /api/sources", sourceHandler.AddSource)

	mux.HandleFunc("POST /api/rooms", roomHandler.CreateRoom)
	mux.HandleFunc("GET /api/rooms/{invite_code}", roomHandler.GetRoom)
	mux.HandleFunc("DELETE /api/rooms/{invite_code}", roomHandler.DeleteRoom)
	mux.HandleFunc("PATCH /api/rooms/{invite_code}/source", roomHandler.PatchRoomSource)

	return &router{
		log:     log,
		Handler: mux,
	}
}

func (r *router) UseLoggingMiddleware() {
	r.Handler = middleware.Logging(r.log, r.Handler)
}

func (r *router) UseAuthMiddleware(sessionStore middleware.SessionStore) {
	r.Handler = middleware.Auth(r.log, r.Handler, sessionStore)
}
