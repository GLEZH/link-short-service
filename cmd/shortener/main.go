package main

import (
	"net/http"
	"os"

	"github.com/GLEZH/linkshrtservice/internal/config"
	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/logger"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		panic(err)
	}

	storage := repository.NewURLStorage()
	handlers := handler.New(cfg.BaseURL, storage)

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	sugar := zapLogger.Sugar()
	sugar.Infow("starting server", "addr", cfg.ServerAddress)

	err = http.ListenAndServe(cfg.ServerAddress, newRouter(handlers, sugar))
	if err != nil {
		panic(err)
	}
}

func newRouter(handlers *handler.Handler, sugar *zap.SugaredLogger) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.Post("/", handlers.ShortenURL)
	router.Get("/{id}", handlers.GetURL)

	return logger.WithLogging(router, sugar)
}
