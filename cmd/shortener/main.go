package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/config"
	"github.com/go-chi/chi/v5"
)

type app struct {
	baseURL string
	mu      sync.RWMutex
	nextID  int
	urls    map[string]string
}

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(cfg.ServerAddress, newRouter(cfg))
	if err != nil {
		panic(err)
	}
}

func newRouter(cfg *config.Config) http.Handler {
	if cfg == nil {
		cfg = &config.Config{
			ServerAddress: ":8080",
			BaseURL:       "http://localhost:8080",
		}
	}

	application := &app{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		urls:    make(map[string]string),
	}

	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.Post("/", application.rootHandler)
	router.Get("/{id}", application.getHandler)

	return router
}

func (a *app) rootHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.nextID++
	id := strconv.Itoa(a.nextID)
	a.urls[id] = originalURL
	a.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(a.baseURL + "/" + id))
}

func (a *app) getHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.mu.RLock()
	originalURL, ok := a.urls[id]
	a.mu.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
