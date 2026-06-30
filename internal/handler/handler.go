package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/go-chi/chi/v5"
)

type URLStorage interface {
	Save(url entity.URL) (entity.URL, error)
	Get(id string) (entity.URL, bool)
}

type Handler struct {
	baseURL string
	storage URLStorage
}

func New(baseURL string, storage URLStorage) *Handler {
	return &Handler{
		baseURL: strings.TrimRight(baseURL, "/"),
		storage: storage,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
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

	shortURL, err := h.storage.Save(entity.URL{OriginalURL: originalURL})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(h.baseURL + "/" + shortURL.ID))
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, ok := h.storage.Get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", shortURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
