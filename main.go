package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type app struct {
	client   *http.Client
	imgbbKey string
	limiter  *rateLimiter
}

func main() {
	_ = godotenv.Load()

	a := &app{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		imgbbKey: os.Getenv("IMGBB_API_KEY"),
		limiter:  newRateLimiter(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.index)
	mux.HandleFunc("POST /upload", a.upload)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := env("ADDR", ":8080")
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, securityHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexHTML)
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
