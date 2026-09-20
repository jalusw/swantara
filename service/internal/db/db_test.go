package db

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/config"
)

func TestNewRejectsInvalidSSLMode(t *testing.T) {
	cfg := &config.Config{
		DatabaseHost:    "localhost",
		DatabasePort:    5432,
		DatabaseUser:    "postgres",
		DatabaseName:    "swantara",
		DatabaseSSLMode: "invalid",
	}

	if _, err := New(cfg); err == nil {
		t.Fatal("New() expected error for invalid sslmode")
	}
}

func TestNewFailsWhenDatabaseUnreachable(t *testing.T) {
	cfg := &config.Config{
		DatabaseHost:          "nonexistent.invalid",
		DatabasePort:          5432,
		DatabaseUser:          "postgres",
		DatabasePassword:      "secret",
		DatabaseName:          "swantara",
		DatabaseSSLMode:       "disable",
		DatabaseMaxOpenConns:  5,
		DatabaseMaxIdleConns:  2,
		DatabaseConnMaxLife:   time.Minute,
		DatabaseConnMaxIdle:   time.Minute,
		DatabaseSlowThreshold: time.Second,
	}

	if _, err := New(cfg); err == nil {
		t.Fatal("New() expected error for unreachable database")
	}
}
