package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

	"w2g/internal/chat"
)

func main() {
	addr := ":" + getEnv("PORT", "8080")

	hub := chat.NewHub(2)
	go hub.Run()

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

	// Проверяем, что web-клиент действительно существует при запуске.
	if _, err := fs.Stat(os.DirFS("."), "web/index.html"); err != nil {
		log.Fatalf("web/index.html not found: %v", err)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("w2g server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
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
