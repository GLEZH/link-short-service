package entity

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidURL reports an empty or invalid URL
	ErrInvalidURL = errors.New("invalid url")
	// ErrURLNotFound reports a missing short URL
	ErrURLNotFound = errors.New("url not found")
	// ErrURLAlreadyExists reports a duplicate original URL
	ErrURLAlreadyExists = errors.New("url already exists")
	// ErrURLDeleted reports a deleted short URL
	ErrURLDeleted = errors.New("url deleted")
	// ErrUserIDNotFound reports a missing user identity
	ErrUserIDNotFound = errors.New("user id not found")
)

// DomainError adds a stable code to a domain error
type DomainError struct {
	message string
	code    string
	err     error
}

// NewDomainError creates a domain error
func NewDomainError(message string, code string, err error) DomainError {
	return DomainError{
		message: message,
		code:    code,
		err:     err,
	}
}

// Error returns the error message
func (e DomainError) Error() string {
	return e.message
}

// Code returns the stable error code
func (e DomainError) Code() string {
	return e.code
}

// Unwrap returns the underlying error
func (e DomainError) Unwrap() error {
	return e.err
}

// InvalidURLError describes an invalid URL
type InvalidURLError struct {
	DomainError
}

// NewInvalidURLError creates an invalid URL error
func NewInvalidURLError() *InvalidURLError {
	return &InvalidURLError{
		DomainError: NewDomainError(
			ErrInvalidURL.Error(),
			"url.invalid",
			ErrInvalidURL,
		),
	}
}

// URLNotFoundError describes a missing short URL
type URLNotFoundError struct {
	DomainError
	ID string
}

// NewURLNotFoundError creates a URL not found error
func NewURLNotFoundError(id string) *URLNotFoundError {
	return &URLNotFoundError{
		DomainError: NewDomainError(
			fmt.Sprintf("%v: id %s", ErrURLNotFound, id),
			"url.not_found",
			ErrURLNotFound,
		),
		ID: id,
	}
}

// URLAlreadyExistsError describes a duplicate original URL
type URLAlreadyExistsError struct {
	DomainError
	URL URL
}

// NewURLAlreadyExistsError creates a duplicate URL error
func NewURLAlreadyExistsError(url URL) *URLAlreadyExistsError {
	return &URLAlreadyExistsError{
		DomainError: NewDomainError(
			fmt.Sprintf("%v: %s", ErrURLAlreadyExists, url.OriginalURL),
			"url.already_exists",
			ErrURLAlreadyExists,
		),
		URL: url,
	}
}

// URLDeletedError describes a deleted short URL
type URLDeletedError struct {
	DomainError
	ID string
}

// NewURLDeletedError creates a deleted URL error
func NewURLDeletedError(id string) *URLDeletedError {
	return &URLDeletedError{
		DomainError: NewDomainError(
			fmt.Sprintf("%v: id %s", ErrURLDeleted, id),
			"url.deleted",
			ErrURLDeleted,
		),
		ID: id,
	}
}

// UserIDNotFoundError describes a missing user identity
type UserIDNotFoundError struct {
	DomainError
}

// NewUserIDNotFoundError creates a missing user ID error
func NewUserIDNotFoundError() *UserIDNotFoundError {
	return &UserIDNotFoundError{
		DomainError: NewDomainError(
			ErrUserIDNotFound.Error(),
			"user.id_not_found",
			ErrUserIDNotFound,
		),
	}
}
