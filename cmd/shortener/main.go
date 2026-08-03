package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GLEZH/linkshrtservice/internal/config"
	"github.com/GLEZH/linkshrtservice/internal/database"
	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/middleware"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer zapLogger.Sync()

	sugar := zapLogger.Sugar()

	db, err := database.New(cfg.DatabaseDSN)
	if err != nil {
		sugar.Fatalw("init database", "error", err)
	}
	defer db.Close()

	storage, closeStorage := newStorage(cfg, db, sugar)
	defer closeStorage()

	handlers := handler.New(cfg.BaseURL, storage, sugar, db)
	sugar.Infow(
		"starting server",
		"addr", cfg.ServerAddress,
		"file_storage_path", cfg.FileStoragePath,
		"database_configured", cfg.DatabaseDSN != "",
	)

	err = http.ListenAndServe(cfg.ServerAddress, newRouter(handlers, sugar))
	if err != nil {
		sugar.Fatalw("start server", "error", err)
	}
}

func newStorage(cfg *config.Config, db *database.DB, sugar *zap.SugaredLogger) (handler.URLStorage, func()) {
	if cfg.DatabaseDSN != "" {
		if err := db.Migrate(); err != nil {
			sugar.Fatalw("run migrations", "error", err)
		}
		return repository.NewDatabaseURLStorage(db.SQLDB()), func() {}
	}

	if cfg.FileStoragePath != "" {
		storage, err := repository.New(cfg.FileStoragePath)
		if err != nil {
			sugar.Fatalw("init file storage", "error", err)
		}
		return storage, func() {
			_ = storage.Close()
		}
	}

	return repository.NewURLStorage(), func() {}
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
	router.Post("/api/shorten", handlers.ShortenURLJSON)
	router.Get("/ping", handlers.PingDB)
	router.Get("/{id}", handlers.GetURL)

	return middleware.WithLogging(middleware.WithGzip(router), sugar)
}
