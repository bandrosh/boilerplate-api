package user

import "errors"

// Domain errors. Adapters (e.g. the HTTP layer) map these to transport-specific
// representations such as status codes, keeping the domain transport-agnostic.
var (
	ErrNameRequired  = errors.New("user: name is required")
	ErrEmailRequired = errors.New("user: email is required")
	ErrInvalidEmail  = errors.New("user: invalid email")
	ErrNotFound      = errors.New("user: not found")
	ErrAlreadyExists = errors.New("user: already exists")
)
