package main

import (
	"net/http"
	"os"

	"github.com/GLEZH/linkshrtservice/internal/config"
	"github.com/go-chi/chi/v5"
)

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

func newRouter(_ *config.Config) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	router.Post("/", rootHandler)
	router.Get("/{id}", getHandler)

	return router
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTemporaryRedirect)
}
