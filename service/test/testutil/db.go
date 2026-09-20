package testutil

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
	"github.com/jalusw/swantara/apps/service/internal/seeders"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

func SystemContext() context.Context {
	return context.Background()
}

func RepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("go.mod not found")
		}
		dir = parent
	}
}

const integrationEnvFile = ".env.test"

func LoadTestConfig() *config.Config {
	cfg, err := config.New(RepoRoot(), integrationEnvFile)
	if err != nil {
		log.Fatalf("failed to load test config: %v", err)
	}
	return cfg
}

func SetupTestDB() *gorm.DB {
	if err := os.Chdir(RepoRoot()); err != nil {
		log.Fatalf("failed to change to repo root: %v", err)
	}

	cfg := LoadTestConfig()

	d, err := db.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}
	if err := d.Use(audit.NewGormPlugin()); err != nil {
		log.Fatalf("failed to register audit plugin: %v", err)
	}

	sqlDB, err := d.DB()
	if err != nil {
		log.Fatalf("failed to open sql handle: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	goose.SetTableName("goose_migrations")

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	seeder, err := seeders.New(d, iam.NewPasswordService(cfg))
	if err != nil {
		log.Fatalf("failed to initialize seeder: %v", err)
	}

	if err := seeder.SeedFoundation(); err != nil {
		log.Fatalf("failed to seed foundation reference data: %v", err)
	}

	if err := seeder.SeedPermissions(); err != nil {
		log.Fatalf("failed to seed permissions: %v", err)
	}

	return d
}
