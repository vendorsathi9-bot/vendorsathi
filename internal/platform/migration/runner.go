// Package migration provides a simple, versioned SQL migration runner that
// uses pgx directly — no third-party migration library required.
//
// Migration files must follow the naming convention:
//
//	<version>_<description>.up.sql   – applied in ascending version order
//	<version>_<description>.down.sql – used for rollback (future use)
//
// Applied migrations are recorded in the schema_migrations table so that
// re-running the runner is always safe (idempotent).
package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT PRIMARY KEY,
    name        TEXT NOT NULL,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// migrationAdvisoryLockID serializes migration runs across API replicas.
// The value is an arbitrary, stable application-specific lock key.
const migrationAdvisoryLockID int64 = 0x56454E4453414154 // "VENDSAAT"

// Migration represents a single versioned SQL migration file.
type Migration struct {
	Version int64
	Name    string
	Path    string
}

// Runner executes SQL migration files against a PostgreSQL database.
type Runner struct {
	pool    *pgxpool.Pool
	migrDir string // path to the directory containing *.up.sql files
}

// NewRunner creates a migration runner.
//
// migrPath may be supplied with or without a "file://" prefix:
//
//	"file://migrations"
//	"migrations"
func NewRunner(pool *pgxpool.Pool, migrPath string) *Runner {
	dir := strings.TrimPrefix(migrPath, "file://")
	return &Runner{pool: pool, migrDir: dir}
}

// Up applies all pending migrations in ascending version order.
// It is safe to call multiple times; already-applied migrations are skipped.
//
// A PostgreSQL advisory lock is held for the duration of the run so that
// multiple application replicas starting at the same time cannot apply the
// same migration concurrently.
func (r *Runner) Up(ctx context.Context) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migration: acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, pgx.QueryExecModeExec, migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("migration: acquire advisory lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.Exec(unlockCtx, `SELECT pg_advisory_unlock($1)`, pgx.QueryExecModeExec, migrationAdvisoryLockID)
	}()

	return r.upLocked(ctx)
}

// upLocked performs the actual migration work. Callers must hold the
// advisory lock returned by Up.
func (r *Runner) upLocked(ctx context.Context) error {
	if err := r.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("migration: ensure schema_migrations table: %w", err)
	}

	migrations, err := r.discoverMigrations()
	if err != nil {
		return fmt.Errorf("migration: discover files: %w", err)
	}
	if len(migrations) == 0 {
		log.Info().Str("dir", r.migrDir).Msg("migration: no migration files found")
		return nil
	}

	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("migration: query applied versions: %w", err)
	}

	pending := filterPending(migrations, applied)
	if len(pending) == 0 {
		log.Info().Msg("migration: all migrations already applied — nothing to do")
		return nil
	}

	for _, m := range pending {
		if err := r.applyOne(ctx, m); err != nil {
			return fmt.Errorf("migration: apply %03d_%s: %w", m.Version, m.Name, err)
		}
	}
	return nil
}

// Status prints the applied/pending state of every migration to stdout.
func (r *Runner) Status(ctx context.Context) error {
	if err := r.ensureMigrationsTable(ctx); err != nil {
		return err
	}
	migrations, err := r.discoverMigrations()
	if err != nil {
		return err
	}
	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		state := "pending"
		if _, ok := applied[m.Version]; ok {
			state = "applied"
		}
		fmt.Printf("  [%s] %03d %s\n", state, m.Version, m.Name)
	}
	return nil
}

// CheckReady verifies that the database schema is fully migrated and that the
// core application tables exist. It is used by the readiness probe so a
// database that is reachable but unmigrated is never reported as ready.
func (r *Runner) CheckReady(ctx context.Context) error {
	migrations, err := r.discoverMigrations()
	if err != nil {
		return fmt.Errorf("discover migrations: %w", err)
	}

	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	var pending []string
	for _, m := range migrations {
		if _, ok := applied[m.Version]; !ok {
			pending = append(pending, fmt.Sprintf("%03d_%s", m.Version, m.Name))
		}
	}
	if len(pending) > 0 {
		return fmt.Errorf("pending migrations: %s", strings.Join(pending, ", "))
	}

	var coreSchemaReady bool
	err = r.pool.QueryRow(ctx,
		`SELECT to_regclass('users') IS NOT NULL AND to_regclass('vendor_profiles') IS NOT NULL`,
		pgx.QueryExecModeSimpleProtocol,
	).Scan(&coreSchemaReady)
	if err != nil {
		return fmt.Errorf("check core schema: %w", err)
	}
	if !coreSchemaReady {
		return fmt.Errorf("core application tables are missing")
	}

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it does not exist.
func (r *Runner) ensureMigrationsTable(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := r.pool.Exec(ctx, createMigrationsTable)
	return err
}

// discoverMigrations scans migrDir for *.up.sql files and parses their versions.
func (r *Runner) discoverMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(r.migrDir)
	if err != nil {
		return nil, fmt.Errorf("read migration directory %q: %w", r.migrDir, err)
	}

	var migrations []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		m, err := parseMigrationFile(filepath.Join(r.migrDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("parse migration filename %q: %w", e.Name(), err)
		}
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

// parseMigrationFile parses a filename like "001_initial_schema.up.sql".
func parseMigrationFile(path string) (Migration, error) {
	base := filepath.Base(path)
	// Strip the .up.sql suffix.
	withoutExt := strings.TrimSuffix(base, ".up.sql")
	parts := strings.SplitN(withoutExt, "_", 2)
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("filename must be <version>_<name>.up.sql, got %q", base)
	}
	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Migration{}, fmt.Errorf("version prefix is not a number in %q", base)
	}
	return Migration{
		Version: version,
		Name:    parts[1],
		Path:    path,
	}, nil
}

// appliedVersions returns a set of version numbers already in schema_migrations.
func (r *Runner) appliedVersions(ctx context.Context) (map[int64]struct{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]struct{})
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = struct{}{}
	}
	return applied, rows.Err()
}

// filterPending returns migrations whose version is not in applied.
func filterPending(all []Migration, applied map[int64]struct{}) []Migration {
	var pending []Migration
	for _, m := range all {
		if _, ok := applied[m.Version]; !ok {
			pending = append(pending, m)
		}
	}
	return pending
}

// applyOne runs a single migration inside a transaction.
// On success it records the version in schema_migrations.
// On failure it rolls back and returns an error — the application should halt.
func (r *Runner) applyOne(ctx context.Context, m Migration) error {
	sql, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("read %s: %w", m.Path, err)
	}

	log.Info().Int64("version", m.Version).Str("name", m.Name).Msg("migration: applying")

	applyCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	tx, err := r.pool.BeginTx(applyCtx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if _, err := tx.Exec(applyCtx, string(sql)); err != nil {
		_ = tx.Rollback(applyCtx)
		return fmt.Errorf("exec SQL: %w", err)
	}

	if _, err := tx.Exec(applyCtx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		pgx.QueryExecModeExec, m.Version, m.Name,
	); err != nil {
		_ = tx.Rollback(applyCtx)
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(applyCtx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Info().Int64("version", m.Version).Str("name", m.Name).Msg("migration: applied successfully")
	return nil
}
