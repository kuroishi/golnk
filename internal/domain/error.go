package domain

import (
	"errors"
)

var (
	ErrInvalidURL = errors.New("invalid URL: must be http or https")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("id conflict")
)