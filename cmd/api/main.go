package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"vendorsaathi/internal/app"
	"vendorsaathi/internal/config"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/platform/logger"
	"vendorsaathi/internal/platform/seeder"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		runSeedCommand()
		return
	}

	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		logger.Logger.Error().Err(err).Msg("application run error")
		os.Exit(1)
	}
}

func runSeedCommand() {
	cfg := config.MustLoad()
	if !cfg.IsDevelopment() {
		fmt.Fprintf(os.Stderr, "Error: Seeding is only permitted in development environments (APP_ENV=development). Current APP_ENV=%q\n", cfg.AppEnv)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel, true)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Check if --reset flag was passed
	for _, arg := range os.Args[2:] {
		if arg == "--reset" {
			if os.Getenv("CONFIRM_DB_RESET") != "yes" {
				fmt.Fprintln(os.Stderr, "Error: Database reset requires explicit confirmation.")
				fmt.Fprintln(os.Stderr, "Set environment variable CONFIRM_DB_RESET=yes to proceed.")
				os.Exit(1)
			}
			fmt.Println("Performing explicit database reset before seeding...")
			if err := seeder.ResetDatabase(ctx, pool); err != nil {
				fmt.Fprintf(os.Stderr, "Reset failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Database reset completed.")
		}
	}

	summary, err := seeder.Seed(ctx, pool, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Seeder error: %v\n", err)
		os.Exit(1)
	}

	summary.Print()
}
