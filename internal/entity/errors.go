package entity

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidURL       = errors.New("invalid url")
	ErrURLNotFound      = errors.New("url not found")
	ErrURLAlreadyExists = errors.New("url already exists")
	ErrURLDeleted       = errors.New("url deleted")
	ErrUserIDNotFound   = errors.New("user id not found")
)

type DomainError struct {
	message string
	code    string
	err     error
}

func NewDomainError(message string, code string, err error) DomainError {
	return DomainError{
		message: message,
		code:    code,
		err:     err,
	}
}

func (e DomainError) Error() string {
	return e.message
}

func (e DomainError) Code() string {
	return e.code
}

func (e DomainError) Unwrap() error {
	return e.err
}

type InvalidURLError struct {
	DomainError
}

func NewInvalidURLError() *InvalidURLError {
	return &InvalidURLError{
		DomainError: NewDomainError(
			ErrInvalidURL.Error(),
			"url.invalid",
			ErrInvalidURL,
		),
	}
}

type URLNotFoundError struct {
	DomainError
	ID string
}

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

type URLAlreadyExistsError struct {
	DomainError
	URL URL
}

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

type URLDeletedError struct {
	DomainError
	ID string
}

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

type UserIDNotFoundError struct {
	DomainError
}

func NewUserIDNotFoundError() *UserIDNotFoundError {
	return &UserIDNotFoundError{
		DomainError: NewDomainError(
			ErrUserIDNotFound.Error(),
			"user.id_not_found",
			ErrUserIDNotFound,
		),
	}
}
