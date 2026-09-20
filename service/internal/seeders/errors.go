package seeders

import "errors"

var (
	ErrDatabaseNotInitialized = errors.New("database is not initialized")
	ErrReferenceMissing       = errors.New("referenced seed data is missing")
)
