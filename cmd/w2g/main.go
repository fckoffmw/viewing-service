package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"w2g/internal/auth"
	"w2g/internal/chat"
	"w2g/internal/middleware"
	"w2g/internal/repo"
	"w2g/internal/room"
	"w2g/internal/source"
)

var log *slog.Logger

func init() {
	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
}

func main() {
	addr := ":" + getEnv("PORT", "8080")

	hub := chat.NewHub(2)
	go hub.Run()

	csvStorage, err := repo.NewCSVStorage("./storage/")
	if err != nil {
		log.Error("when creating csv storage", "error", err)
		os.Exit(1)
	}

	sourceService := source.NewService(csvStorage)
	sourceHandler := source.NewHandler(sourceService, log)

	authService := auth.NewService(csvStorage)
	authHandler := auth.NewHandler(authService)

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
	mux.HandleFunc("POST /api/login", authHandler.Login)
	// logout
	// register

	// sources
	mux.HandleFunc("GET /api/sources", sourceHandler.GetAllSources)

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

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if key == "PORT" {
		if _, err := strconv.Atoi(value); err != nil {
			return fallback
		}
	}
	return value
}
