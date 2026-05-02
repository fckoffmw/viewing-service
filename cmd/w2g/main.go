package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"w2g/internal/auth"
	"w2g/internal/chat"
	"w2g/internal/config"
	router "w2g/internal/http"
	"w2g/internal/repo"
	"w2g/internal/room"
	"w2g/internal/source"
)

var log *slog.Logger

func main() {
	config := config.Load()

	initLogger(config.LogLevel, config.LogFile)

	log.Info("config", "config", config.PrettyPrint())

	csvStorage, err := repo.NewCSVStorage(config.StorageDir)
	if err != nil {
		log.Error("when creating csv storage", "error", err)
		os.Exit(1)
	}

	roomStore, err := room.NewStore(csvStorage)
	if err != nil {
		log.Error("when creating room store", "error", err)
		os.Exit(1)
	}

	sessionStore := auth.NewSessionStore(
		time.Duration(config.SessionsCleanupInterval) * time.Second,
	)
	go sessionStore.CleanupLoop()

	sourceService := source.NewService(csvStorage)
	sourceHandler := source.NewHandler(sourceService, log)

	authService := auth.NewService(csvStorage, sessionStore)
	authHandler := auth.NewHandler(authService, log)

	roomService := room.NewService(roomStore, config.MaxRoomsPerUser)
	roomHubManager := chat.NewHubManager()
	roomHandler := room.NewHandler(roomService, roomHubManager, csvStorage, log)

	hub := chat.NewHub(config.MaxClients)
	go hub.Run()

	r := router.NewRouter(log, hub, authService, authHandler, sourceHandler, roomHandler, roomHubManager)

	r.UseLoggingMiddleware()
	r.UseAuthMiddleware(sessionStore)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: r.Handler,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("w2g server listening", "port", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
		}
	}()

	<-sigCh

	shutdownTimeout := 30 * time.Second
	log.Info("shutting down...", "timeout", shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Info("stopping HTTP server...")
	server.Shutdown(ctx)

	log.Info("stopping session cleanup...")
	sessionStore.Stop()

	log.Info("stopping chat hub...")
	hub.Close()

	log.Info("shutdown complete")
}

func initLogger(level, logFile string) {
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

	var handlers []slog.Handler

	handlers = append(handlers, slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelValue,
	}))

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			slog.Error("cannot open log file", "path", logFile, "err", err)
		} else {
			handlers = append(handlers, slog.NewTextHandler(f, &slog.HandlerOptions{
				Level: levelValue,
			}))
		}
	}

	log = slog.New(slog.NewMultiHandler(handlers...))
}
