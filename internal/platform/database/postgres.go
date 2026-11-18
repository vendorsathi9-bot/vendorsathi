package database

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool = pgxpool.Pool

// DBTX is a shared database executor interface implemented by both *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// DBFromContext returns the pgx.Tx if stored in ctx by TxManager.RunInTx; otherwise fallback.
func DBFromContext(ctx context.Context, fallback DBTX) DBTX {
	if tx, ok := GetPgxTx(ctx); ok {
		return tx
	}
	return fallback
}

type Config struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	poolConfig.ConnConfig.RuntimeParams = map[string]string{
		"application_name": "vendorsaathi-api",
		"timezone":         "UTC",
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func DefaultConfig(url string) Config {
	return Config{
		URL:               url,
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// IsRetryableTxErr reports whether an error represents a PostgreSQL deadlock or serialization failure
// that is safe to retry with exponential backoff.
func IsRetryableTxErr(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 40P01 = deadlock_detected, 40001 = serialization_failure
		return pgErr.Code == "40P01" || pgErr.Code == "40001"
	}
	return false
}

func (tm *TxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return tm.RunInTxWithRetry(ctx, 3, fn)
}

// RunInTxWithRetry executes fn inside a transaction, retrying up to maxAttempts on deadlock or serialization errors.
func (tm *TxManager) RunInTxWithRetry(ctx context.Context, maxAttempts int, fn func(context.Context) error) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(25*(1<<attempt)) * time.Millisecond
			jitter := time.Duration(rand.Intn(20)) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}

		tx, err := tm.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		err = func() (txErr error) {
			defer func() {
				if p := recover(); p != nil {
					_ = tx.Rollback(ctx)
					panic(p)
				}
			}()

			txCtx := context.WithValue(ctx, txKey{}, tx)
			if fnErr := fn(txCtx); fnErr != nil {
				_ = tx.Rollback(ctx)
				return fnErr
			}

			if commitErr := tx.Commit(ctx); commitErr != nil {
				return fmt.Errorf("commit transaction: %w", commitErr)
			}
			return nil
		}()

		if err == nil {
			return nil
		}

		lastErr = err
		if !IsRetryableTxErr(err) || ctx.Err() != nil {
			return err
		}
	}
	return lastErr
}

func (tm *TxManager) Pool() *pgxpool.Pool {
	return tm.pool
}

type txKey struct{}

// GetPgxTx retrieves a pgx.Tx stored in the context by TxManager.RunInTx.
// Repositories that want to participate in a transaction should call this
// and fall back to their pool when no transaction is present.
func GetPgxTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
