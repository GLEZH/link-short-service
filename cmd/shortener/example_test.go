package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"go.uber.org/zap"
)

func exampleRouter() http.Handler {
	log := zap.NewNop().Sugar()
	handlers := handler.New(
		"http://localhost:8080",
		repository.NewURLStorage(),
		log,
		nil,
	)
	return newRouter(handlers, log, auth.NewManager("example-secret"))
}

func Example_textEndpoints() {
	router := exampleRouter()

	shortenRequest := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://example.com/articles/42"),
	)
	shortenResponse := httptest.NewRecorder()
	router.ServeHTTP(shortenResponse, shortenRequest)

	fmt.Println(shortenResponse.Code)
	fmt.Println(shortenResponse.Body.String())

	followRequest := httptest.NewRequest(http.MethodGet, "/1", nil)
	followResponse := httptest.NewRecorder()
	router.ServeHTTP(followResponse, followRequest)

	fmt.Println(followResponse.Code)
	fmt.Println(followResponse.Header().Get("Location"))

	// Output:
	// 201
	// http://localhost:8080/1
	// 307
	// https://example.com/articles/42
}

func Example_jsonEndpoint() {
	router := exampleRouter()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(`{"url":"https://example.com/articles/42"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	fmt.Println(response.Code)
	fmt.Print(response.Body.String())

	// Output:
	// 201
	// {"result":"http://localhost:8080/1"}
}

func Example_batchEndpoint() {
	router := exampleRouter()
	body := `[
		{"correlation_id":"first","original_url":"https://example.com/first"},
		{"correlation_id":"second","original_url":"https://example.com/second"}
	]`
	request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	fmt.Println(response.Code)
	fmt.Print(response.Body.String())

	// Output:
	// 201
	// [{"correlation_id":"first","short_url":"http://localhost:8080/1"},{"correlation_id":"second","short_url":"http://localhost:8080/2"}]
}
