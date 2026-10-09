package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GLEZH/linkshrtservice/internal/audit"
	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/config"
	"github.com/GLEZH/linkshrtservice/internal/database"
	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/middleware"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const shutdownTimeout = 10 * time.Second

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
	if db != nil {
		defer db.Close()
	}

	storage, closeStorage := newStorage(cfg, db, sugar)
	defer closeStorage()

	auditPublisher := audit.NewPublisher(sugar)
	defer func() {
		if err := auditPublisher.Close(); err != nil {
			sugar.Errorw("close audit publisher", "error", err)
		}
	}()

	if cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			sugar.Fatalw("init audit file", "error", err)
		}
		auditPublisher.Subscribe(fileObserver)
	}
	if cfg.AuditURL != "" {
		auditPublisher.Subscribe(audit.NewHTTPObserver(cfg.AuditURL, nil))
	}

	handlers := handler.New(cfg.BaseURL, storage, sugar, db, auditPublisher)
	defer func() {
		if err := handlers.Close(); err != nil {
			sugar.Errorw("close handlers", "error", err)
		}
	}()

	sugar.Infow(
		"starting server",
		"addr", cfg.ServerAddress,
		"file_storage_path", cfg.FileStoragePath,
		"database_configured", cfg.DatabaseDSN != "",
	)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: newRouter(handlers, sugar, auth.NewManager(cfg.AuthSecret)),
	}
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err = runServer(shutdownCtx, server); err != nil {
		sugar.Errorw("run server", "error", err)
	}
}

func runServer(ctx context.Context, server *http.Server) error {
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		err := <-serverErrors
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
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

func newRouter(handlers *handler.Handler, sugar *zap.SugaredLogger, authManager *auth.Manager) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.Post("/", handlers.ShortenURL)
	router.Post("/api/shorten", handlers.ShortenURLJSON)
	router.Post("/api/shorten/batch", handlers.ShortenURLBatch)
	router.Get("/api/user/urls", handlers.GetUserURLs)
	router.Delete("/api/user/urls", handlers.DeleteUserURLs)
	router.Get("/ping", handlers.PingDB)
	router.Get("/{id}", handlers.GetURL)

	return middleware.WithLogging(middleware.WithGzip(authManager.WithAuth(router)), sugar)
}
