package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	err := http.ListenAndServe(":8080", newRouter())
	if err != nil {
		panic(err)
	}
}

func newRouter() http.Handler {
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
