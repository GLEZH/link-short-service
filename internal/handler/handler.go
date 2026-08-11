package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLStorage interface {
	Save(ctx context.Context, url entity.URL) (entity.URL, error)
	SaveBatch(ctx context.Context, urls []entity.URL) ([]entity.URL, error)
	Get(ctx context.Context, id string) (entity.URL, error)
	GetByUserID(ctx context.Context, userID string) ([]entity.URL, error)
}

type Database interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	baseURL string
	storage URLStorage
	log     *zap.SugaredLogger
	db      Database
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

type shortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type shortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type userURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(baseURL string, storage URLStorage, log *zap.SugaredLogger, db Database) *Handler {
	return &Handler{
		baseURL: strings.TrimRight(baseURL, "/"),
		storage: storage,
		log:     log,
		db:      db,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.createShortURL(r.Context(), string(body))
	if err != nil {
		if errors.Is(err, entity.ErrURLAlreadyExists) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(shortURL))
			return
		}
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

	shortURL, err := h.createShortURL(r.Context(), request.URL)
	if err != nil {
		if errors.Is(err, entity.ErrURLAlreadyExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(shortenResponse{Result: shortURL})
			return
		}
		h.writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{Result: shortURL})
}

func (h *Handler) ShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	var request []shortenBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(request) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urls := make([]entity.URL, 0, len(request))
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	for _, item := range request {
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			h.writeError(w, entity.NewInvalidURLError())
			return
		}
		urls = append(urls, entity.URL{OriginalURL: originalURL, UserID: userID})
	}

	savedURLs, err := h.storage.SaveBatch(r.Context(), urls)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := make([]shortenBatchResponse, 0, len(savedURLs))
	for i, savedURL := range savedURLs {
		response = append(response, shortenBatchResponse{
			CorrelationID: request[i].CorrelationID,
			ShortURL:      h.baseURL + "/" + savedURL.ID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.storage.GetByUserID(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]userURLResponse, 0, len(urls))
	for _, url := range urls {
		response = append(response, userURLResponse{
			ShortURL:    h.baseURL + "/" + url.ID,
			OriginalURL: url.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.Header().Set("Location", shortURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.db.Ping(r.Context()); err != nil {
		if h.log != nil {
			h.log.Infow("database ping failed", "error", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createShortURL(ctx context.Context, originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", entity.NewInvalidURLError()
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return "", entity.NewUserIDNotFoundError()
	}

	shortURL, err := h.storage.Save(ctx, entity.URL{OriginalURL: originalURL, UserID: userID})
	if err != nil {
		var alreadyExists *entity.URLAlreadyExistsError
		if errors.As(err, &alreadyExists) {
			return h.baseURL + "/" + alreadyExists.URL.ID, err
		}
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
	case errors.Is(err, entity.ErrUserIDNotFound):
		w.WriteHeader(http.StatusUnauthorized)
	default:
		if h.log != nil {
			h.log.Infow("internal handler error", "error", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}
