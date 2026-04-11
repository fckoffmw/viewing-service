package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"w2g/internal/auth"
	"w2g/internal/chat"
	"w2g/internal/frame"
	"w2g/internal/repo"
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

	frameService := frame.NewService(csvStorage)
	frameHandler := frame.NewHandler(frameService)

	authService := auth.NewService(csvStorage)
	authHandler := auth.NewHandler(authService)

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWS(hub, w, r)
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

	// frames
	mux.HandleFunc("GET /api/frames", frameHandler.GetAllFrames)

	if _, err := fs.Stat(os.DirFS("."), "web/index.html"); err != nil {
		log.Error("web/index.html not found", "error", err)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Info("w2g server listening", "port", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("error", err)
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
