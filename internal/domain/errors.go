// internal/domain/errors.go
package domain

import "errors"

var (
    ErrUserNotFound   = errors.New("user not found")
    ErrInvalidName    = errors.New("invalid name")
    ErrInvalidEmail   = errors.New("invalid email")
    ErrInvalidAge     = errors.New("invalid age")
    ErrEmailExists    = errors.New("email already exists")
)