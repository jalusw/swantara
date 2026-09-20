package xtradata

import "errors"

var (
	ErrEventNotFound    = errors.New("integration event not found")
	ErrConfigKeyExists  = errors.New("system config key already exists")
	ErrEventAlreadySent = errors.New("integration event already sent")
	ErrIdempotencySave  = errors.New("failed to store idempotency key")
)
