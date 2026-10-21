package entity

import (
	"errors"
	"testing"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name string
		err  interface {
			error
			Code() string
			Unwrap() error
		}
		wantMessage string
		wantCode    string
		wantCause   error
	}{
		{
			name:        "invalid URL",
			err:         NewInvalidURLError(),
			wantMessage: "invalid url",
			wantCode:    "url.invalid",
			wantCause:   ErrInvalidURL,
		},
		{
			name:        "URL not found",
			err:         NewURLNotFoundError("short-id"),
			wantMessage: "url not found: id short-id",
			wantCode:    "url.not_found",
			wantCause:   ErrURLNotFound,
		},
		{
			name: "URL already exists",
			err: NewURLAlreadyExistsError(URL{
				OriginalURL: "https://example.com",
			}),
			wantMessage: "url already exists: https://example.com",
			wantCode:    "url.already_exists",
			wantCause:   ErrURLAlreadyExists,
		},
		{
			name:        "URL deleted",
			err:         NewURLDeletedError("short-id"),
			wantMessage: "url deleted: id short-id",
			wantCode:    "url.deleted",
			wantCause:   ErrURLDeleted,
		},
		{
			name:        "user ID not found",
			err:         NewUserIDNotFoundError(),
			wantMessage: "user id not found",
			wantCode:    "user.id_not_found",
			wantCause:   ErrUserIDNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err.Error() != test.wantMessage {
				t.Errorf("Error() = %q, want %q", test.err.Error(), test.wantMessage)
			}
			if test.err.Code() != test.wantCode {
				t.Errorf("Code() = %q, want %q", test.err.Code(), test.wantCode)
			}
			if !errors.Is(test.err, test.wantCause) {
				t.Errorf("error does not wrap %v", test.wantCause)
			}
			if test.err.Unwrap() != test.wantCause {
				t.Errorf("Unwrap() = %v, want %v", test.err.Unwrap(), test.wantCause)
			}
		})
	}
}
