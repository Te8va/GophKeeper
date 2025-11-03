package errors

import "errors"

var (
	ErrURLExists = errors.New("URL уже существует")
	ErrDeleted   = errors.New("удалено")
	ErrNotFound  = errors.New("URL не найден")
)

var (
	ErrAlreadyRegistered = errors.New("user with this username is already registered")
	ErrUserNotFound      = errors.New("user not found")
	ErrWrongPassword     = errors.New("wrong password provided")
	ErrTooManyRequests   = errors.New("too many requests to accrual system")
	ErrInvalidInput      = errors.New("invalid type")
	ErrNotFoundData      = errors.New("data not found")
	ErrInvalidId         = errors.New("missing or invalid ID")
)
