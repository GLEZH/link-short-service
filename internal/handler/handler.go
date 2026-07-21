package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLStorage interface {
	Save(url entity.URL) (entity.URL, error)
	Get(id string) (entity.URL, error)
}

type Handler struct {
	baseURL string
	storage URLStorage
	log     *zap.SugaredLogger
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func New(baseURL string, storage URLStorage, log *zap.SugaredLogger) *Handler {
	return &Handler{
		baseURL: strings.TrimRight(baseURL, "/"),
		storage: storage,
		log:     log,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.createShortURL(string(body))
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

func (h *Handler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var request shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.createShortURL(request.URL)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{Result: shortURL})
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Get(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Location", shortURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) createShortURL(originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", entity.ErrInvalidURL
	}

	shortURL, err := h.storage.Save(entity.URL{OriginalURL: originalURL})
	if err != nil {
		return "", err
	}

	return h.baseURL + "/" + shortURL.ID, nil
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, entity.ErrInvalidURL):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, entity.ErrURLNotFound):
		w.WriteHeader(http.StatusBadRequest)
	default:
		if h.log != nil {
			h.log.Infow("internal handler error", "error", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}
