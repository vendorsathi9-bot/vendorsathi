// cmd/migrate is a standalone migration CLI for VendorSaathi.
//
// Usage:
//
//	go run ./cmd/migrate up      – apply all pending migrations
//	go run ./cmd/migrate status  – show applied/pending state of each migration
//
// Configuration is read the same way as the API: config.yaml (if present)
// and/or environment variables.
//
// Required environment variables (or config.yaml equivalents):
//
//	DATABASE_URL   – PostgreSQL connection string
//	MIGRATION_PATH – Directory containing *.up.sql files (default: migrations)
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"vendorsaathi/internal/config"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/platform/logger"
	"vendorsaathi/internal/platform/migration"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "up", "status":
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", command)
		printUsage()
		os.Exit(1)
	}

	cfg := config.MustLoad()
	logger.Init(cfg.LogLevel, cfg.IsDevelopment())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration: failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	runner := migration.NewRunner(pool, cfg.Migration.Path)

	switch command {
	case "up":
		if err := runner.Up(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migration: all migrations applied successfully")
	case "status":
		if err := runner.Status(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "migration status error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/migrate <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  up      Apply all pending migrations")
	fmt.Fprintln(os.Stderr, "  status  Show applied/pending migration status")
}
