package domain

import "errors"

// Domain errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("conflict")
	ErrInternal          = errors.New("internal error")
	ErrValidation        = errors.New("validation error")
	ErrDraftNotFound     = errors.New("draft version not found")
	ErrActiveNotFound    = errors.New("active version not found")
	ErrVersionConflict   = errors.New("version conflict")
	ErrMaxVersionReached = errors.New("maximum version reached")
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidVersion    = errors.New("invalid version")
)
