package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	connectAttempts = 5
	connectBackoff  = 2 * time.Second

	pgUniqueViolation = "23505"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

func New(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%v",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)

	pgConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection: %w", err)
	}

	sqlDB := sql.OpenDB(stdlib.GetConnector(*pgConfig))

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: newGormSlogLogger(cfg.DatabaseSlowThreshold),
	})
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DatabaseConnMaxLife)
	sqlDB.SetConnMaxIdleTime(cfg.DatabaseConnMaxIdle)

	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := sqlDB.PingContext(ctx)
		cancel()

		if err == nil {
			break
		}

		if attempt >= connectAttempts {
			return nil, fmt.Errorf("failed to ping database after %d attempts: %w", attempt, err)
		}

		slog.Warn("database not ready, retrying", "attempt", attempt, "max_attempts", connectAttempts, "error", err)
		time.Sleep(connectBackoff)
	}

	return db, nil
}
