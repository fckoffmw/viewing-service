package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func main() {
	dir := flag.String("dir", "./web", "directory to serve")
	port := flag.String("port", "8081", "port to listen on")
	backend := flag.String("backend", "http://localhost:8080", "backend URL to proxy API requests")
	flag.Parse()

	addr := ":" + *port
	log.Printf("serving %s on %s", *dir, addr)
	log.Printf("proxying /auth/*, /api/*, /ws -> %s", *backend)

	backendURL, err := url.Parse(*backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid backend URL: %v\n", err)
		os.Exit(1)
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		if path == "/auth/me" || path == "/auth/login" || path == "/auth/logout" ||
			path == "/auth/register" || path == "/api/sources" ||
			path == "/healthz" || strings.HasPrefix(path, "/ws/") ||
			(len(path) > 5 && path[:5] == "/api/") {

			log.Printf("%s %s -> proxy", r.Method, path)
			proxy.ServeHTTP(w, r)
			log.Printf("  -> proxied in %v", time.Since(start))
			return
		}

		log.Printf("%s %s", r.Method, path)

		wrapped := &statusWriter{ResponseWriter: w, status: 200}
		http.FileServer(http.Dir(*dir)).ServeHTTP(wrapped, r)

		log.Printf("  -> %d in %v", wrapped.status, time.Since(start))
	})

	log.Printf("open http://localhost:%s", *port)

	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}