package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "post root works",
			method:         http.MethodPost,
			path:           "/",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "get root is bad request",
			method:         http.MethodGet,
			path:           "/",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "post wrong path is bad request",
			method:         http.MethodPost,
			path:           "/test",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			recorder := httptest.NewRecorder()

			rootHandler(recorder, request)

			if recorder.Code != test.wantStatusCode {
				t.Errorf("status code = %d, want %d", recorder.Code, test.wantStatusCode)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "get id works",
			method:         http.MethodGet,
			path:           "/abc123",
			wantStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:           "post id is bad request",
			method:         http.MethodPost,
			path:           "/abc123",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty id is bad request",
			method:         http.MethodGet,
			path:           "/",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			recorder := httptest.NewRecorder()

			getHandler(recorder, request)

			if recorder.Code != test.wantStatusCode {
				t.Errorf("status code = %d, want %d", recorder.Code, test.wantStatusCode)
			}
		})
	}
}
