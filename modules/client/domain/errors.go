package domain

import "errors"

var (
	ErrClientNotFound     = errors.New("client not found")
	ErrEmailAlreadyExists = errors.New("email already registered")
)
