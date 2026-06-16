package repository

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateUsername = errors.New("username already taken")
	ErrDuplicateProvider = errors.New("provider already linked")
)
