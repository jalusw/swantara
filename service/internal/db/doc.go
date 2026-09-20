// Package db bootstraps the GORM/Postgres connection, provides transaction
// plumbing through the Transactioner abstraction, and wires a slog-based GORM
// logger for query observability.
package db
