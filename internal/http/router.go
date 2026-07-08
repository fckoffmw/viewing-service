package http

import (
	"log/slog"
	"net/http"

	"w2g/internal/auth"
	"w2g/internal/realtime"
	"w2g/internal/middleware"
)

type authHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
}

type sourceHandler interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	Add(w http.ResponseWriter, r *http.Request)
	Patch(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type roomHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	PatchSource(w http.ResponseWriter, r *http.Request)
}

type router struct {
	log     *slog.Logger
	Handler http.Handler
}

func NewRouter(
	log *slog.Logger,
	authService auth.Service,
	authHandler authHandler,
	sourceHandler sourceHandler,
	roomHandler roomHandler,
	hubManager realtime.HubGetter,
) *router {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws/{invite_code}", func(w http.ResponseWriter, r *http.Request) {
		realtime.ServeWS(log, hubManager, authService, w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /auth/me", authHandler.Me)

	mux.HandleFunc("GET /api/sources", sourceHandler.GetAll)
	mux.HandleFunc("POST /api/sources", sourceHandler.Add)
	mux.HandleFunc("PATCH /api/sources/{id}", sourceHandler.Patch)
	mux.HandleFunc("DELETE /api/sources/{id}", sourceHandler.Delete)

	mux.HandleFunc("POST /api/rooms", roomHandler.Create)
	mux.HandleFunc("GET /api/rooms", roomHandler.GetAll)
	mux.HandleFunc("GET /api/rooms/{invite_code}", roomHandler.Get)
	mux.HandleFunc("DELETE /api/rooms/{invite_code}", roomHandler.Delete)
	mux.HandleFunc("PATCH /api/rooms/{invite_code}/source", roomHandler.PatchSource)

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
