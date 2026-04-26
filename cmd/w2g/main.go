package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"w2g/internal/auth"
	"w2g/internal/chat"
	"w2g/internal/config"
	"w2g/internal/middleware"
	"w2g/internal/repo"
	"w2g/internal/room"
	"w2g/internal/source"
)

var log *slog.Logger

func main() {

	config := config.Load()

	addr := ":" + config.Port

	hub := chat.NewHub(config.MaxClients)
	go hub.Run()

	initLogger(config.LogLevel)

	log.Info("config",
		"STORAGE_DIR", config.StorageDir,
		"PORT", config.Port,
		"MAX_CLIENTS", config.MaxClients,
		"LOG_LEVEL", config.LogLevel,
	)

	csvStorage, err := repo.NewCSVStorage(config.StorageDir)
	if err != nil {
		log.Error("when creating csv storage", "error", err)
		os.Exit(1)
	}

	sessionStore := auth.NewSessionStore()

	sourceService := source.NewService(csvStorage)
	sourceHandler := source.NewHandler(sourceService, log)

	authService := auth.NewService(csvStorage, sessionStore)
	authHandler := auth.NewHandler(authService, log)

	roomService := room.NewService(csvStorage)
	roomHandler := room.NewHandler(roomService, log)

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWS(log, hub, w, r)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// api

	// login
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	// logout
	// register
	mux.HandleFunc("POST /auth/register", authHandler.Register)

	// sources
	mux.HandleFunc("GET /api/sources", sourceHandler.GetAllSources)
	mux.HandleFunc("POST /api/sources", sourceHandler.AddSource)

	// room
	mux.HandleFunc("GET /api/room", roomHandler.GetGlobalRoom)
	mux.HandleFunc("PATCH /api/room/source", roomHandler.PatchGlobalRoomSource)

	if _, err := fs.Stat(os.DirFS("."), "web/index.html"); err != nil {
		log.Error("web/index.html not found", "error", err)
	}

	wrappedMux := middleware.Logging(log, mux)

	server := &http.Server{
		Addr:    addr,
		Handler: wrappedMux,
	}

	log.Info("w2g server listening", "port", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "err", err)
	}
}

func initLogger(level string) {
	var levelValue slog.Level
	switch strings.ToLower(level) {
	case "info":
		levelValue = slog.LevelInfo
	case "warn":
		levelValue = slog.LevelWarn
	case "error":
		levelValue = slog.LevelError
	default:
		levelValue = slog.LevelDebug
	}

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: levelValue,
		}),
	)
}
