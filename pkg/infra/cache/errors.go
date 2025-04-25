package cache

import "errors"

// Cache-specific errors
var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrKeyExists       = errors.New("key already exists")
	ErrLockAcquired    = errors.New("failed to acquire lock")
	ErrLockExpired     = errors.New("lock expired")
	ErrConnectionFail  = errors.New("connection failed")
	ErrInvalidArgument = errors.New("invalid argument")
)
