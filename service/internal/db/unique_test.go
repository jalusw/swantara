package db

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Error("expected true for 23505")
	}
	if IsUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Error("expected false for 23503")
	}
	if IsUniqueViolation(errors.New("boom")) {
		t.Error("expected false for generic error")
	}
	if IsUniqueViolation(nil) {
		t.Error("expected false for nil")
	}
}
