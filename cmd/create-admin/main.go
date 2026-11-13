// cmd/create-admin securely bootstraps the first production administrator.
//
// It reads credentials from the environment so they never appear in shell
// history or process listings:
//
//	ADMIN_PHONE=+919876543210
//	ADMIN_PASSWORD='a-strong-unique-password'
//	ADMIN_EMAIL=admin@example.com          # optional
//	ADMIN_UPDATE_PASSWORD=true             # optional, existing user only
//
// The account is created with ACTIVE status and the SUPER_ADMIN role. If the
// user already exists, the role is ensured; the password is only changed when
// ADMIN_UPDATE_PASSWORD=true is explicitly set.
//
// Run migrations first (the API does this automatically at startup):
//
//	go run ./cmd/create-admin
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"vendorsaathi/internal/config"
	authmodel "vendorsaathi/internal/modules/auth/model"
	authrepository "vendorsaathi/internal/modules/auth/repository"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/platform/logger"
	"vendorsaathi/internal/shared"
)

var e164PhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "create-admin: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger.Init(cfg.LogLevel, cfg.IsDevelopment())

	phone := strings.TrimSpace(os.Getenv("ADMIN_PHONE"))
	password := os.Getenv("ADMIN_PASSWORD")
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	updatePassword := os.Getenv("ADMIN_UPDATE_PASSWORD") == "true"

	if err := validateBootstrapInput(phone, password, email); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	repo := authrepository.NewUserRepository(pool)
	txManager := database.NewTxManager(pool)

	passwordHash, err := auth.HashPassword(
		password,
		cfg.PasswordHash.Memory,
		cfg.PasswordHash.Iterations,
		cfg.PasswordHash.Parallelism,
		cfg.PasswordHash.SaltLength,
		cfg.PasswordHash.KeyLength,
	)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	var userID uuid.UUID
	var created bool

	txErr := txManager.RunInTx(ctx, func(txCtx context.Context) error {
		existing, lookupErr := repo.GetUserByPhone(txCtx, phone)
		switch {
		case lookupErr == nil && existing != nil:
			if existing.Status != authmodel.UserStatusActive {
				return fmt.Errorf(
					"existing user %s has status %s; reactivate it intentionally before granting admin access",
					phone, existing.Status,
				)
			}
			if updatePassword {
				if err := repo.UpdatePassword(txCtx, existing.ID, passwordHash); err != nil {
					return fmt.Errorf("update password: %w", err)
				}
			}
			userID = existing.ID

		case errors.Is(lookupErr, shared.ErrNotFound):
			user, createErr := repo.CreateUserWithStatus(
				txCtx, phone, email, passwordHash, authmodel.UserStatusActive,
			)
			if createErr != nil {
				return fmt.Errorf("create user: %w", createErr)
			}
			userID = user.ID
			created = true

		default:
			return fmt.Errorf("look up user: %w", lookupErr)
		}

		role, roleErr := repo.GetRoleByCode(txCtx, "SUPER_ADMIN")
		if roleErr != nil {
			return fmt.Errorf("load SUPER_ADMIN role (have migrations been applied?): %w", roleErr)
		}
		if err := repo.AssignRole(txCtx, userID, role.ID); err != nil {
			return fmt.Errorf("assign SUPER_ADMIN role: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}

	if created {
		fmt.Printf("created SUPER_ADMIN user %s (%s)\n", phone, userID)
	} else {
		fmt.Printf("ensured SUPER_ADMIN role for existing user %s (%s)\n", phone, userID)
	}
	return nil
}

func validateBootstrapInput(phone, password, email string) error {
	if phone == "" {
		return errors.New("ADMIN_PHONE is required (E.164 format, e.g. +919876543210)")
	}
	if !e164PhonePattern.MatchString(phone) {
		return errors.New("ADMIN_PHONE must be in E.164 format, e.g. +919876543210")
	}
	if password == "" {
		return errors.New("ADMIN_PASSWORD is required")
	}
	if len(password) < 12 {
		return errors.New("ADMIN_PASSWORD must be at least 12 characters")
	}
	switch password {
	case "AdminPassword123!", "VendorPassword123!", "CustomerPassword123!":
		return errors.New("ADMIN_PASSWORD must not be a documented demo password")
	}
	if email != "" && (!strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@")) {
		return errors.New("ADMIN_EMAIL must be a valid email address")
	}
	return nil
}
