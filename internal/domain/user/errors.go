package user

import "errors"

var (
	ErrNameRequired  = errors.New("user: name is required")
	ErrEmailRequired = errors.New("user: email is required")
	ErrInvalidEmail  = errors.New("user: invalid email")
	ErrNotFound      = errors.New("user: not found")
	ErrAlreadyExists = errors.New("user: already exists")
)
