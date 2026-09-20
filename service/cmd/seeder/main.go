package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/logger"
	"github.com/jalusw/swantara/apps/service/internal/seeders"
)

func main() {
	startEstimation := time.Now()

	cfg, err := config.New("", config.ConfigFilePath())
	if err != nil {
		logger.Fatalf("Failed to load config %v", err)
	}

	logger.Setup(cfg.ApplicationDebug, cfg.LogLevel, logger.Options{
		File:       cfg.LogFile,
		Format:     cfg.LogFormat,
		MaxSize:    cfg.LogMaxSize,
		MaxAge:     cfg.LogMaxAge,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
	})
	slog.Info("Starting seeding process")

	d, err := db.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to load config %v", err)
	}

	passwordSvc := iam.NewPasswordService(cfg)
	seeder, err := seeders.New(d, passwordSvc)
	if err != nil {
		logger.Fatalf("Failed to initialize seeder %v", err)
	}

	if err := seeder.SeedFoundation(); err != nil {
		logger.Fatalf("Failed to seed foundation reference data %v", err)
	}
	if err := seeder.SeedPermissions(); err != nil {
		logger.Fatalf("Failed to seed permissions %v", err)
	}
	if err := seeder.SeedDemoUsers(); err != nil {
		logger.Fatalf("Failed to seed demo users %v", err)
	}

	elapsed := time.Since(startEstimation)
	fmt.Printf("Seeding finished with %v", elapsed.Milliseconds())
}
