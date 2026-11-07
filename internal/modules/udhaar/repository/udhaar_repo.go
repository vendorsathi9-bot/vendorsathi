package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/udhaar/model"
	"vendorsaathi/internal/shared"
)

type UdhaarRepository struct {
	pool *pgxpool.Pool
}

func NewUdhaarRepository(pool *pgxpool.Pool) *UdhaarRepository {
	return &UdhaarRepository{pool: pool}
}

// CreateAccount inserts a new udhaar account for a vendor-customer pair.
func (r *UdhaarRepository) CreateAccount(ctx context.Context, vendorID, customerID uuid.UUID, creditLimitPaise *int64) (*model.UdhaarAccount, error) {
	id := uuid.New()
	var a model.UdhaarAccount
	err := r.pool.QueryRow(ctx, `
		INSERT INTO udhaar_accounts (id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at)
		VALUES ($1, $2, $3, 'ACTIVE', $4, NOW(), NOW())
		RETURNING id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
	`, id, vendorID, customerID, creditLimitPaise,
	).Scan(&a.ID, &a.VendorID, &a.CustomerID, &a.Status, &a.CreditLimitPaise, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, shared.NewConflict("udhaar account already exists for this vendor-customer pair")
		}
		return nil, err
	}
	return &a, nil
}

// GetAccountByID fetches a single udhaar account by primary key.
func (r *UdhaarRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*model.UdhaarAccount, error) {
	var a model.UdhaarAccount
	err := r.pool.QueryRow(ctx, `
		SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
		FROM udhaar_accounts WHERE id = $1
	`, id).Scan(&a.ID, &a.VendorID, &a.CustomerID, &a.Status, &a.CreditLimitPaise, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("udhaar account")
		}
		return nil, err
	}
	return &a, nil
}

// GetAccountByVendorAndCustomer fetches the account for a specific vendor-customer pair.
func (r *UdhaarRepository) GetAccountByVendorAndCustomer(ctx context.Context, vendorID, customerID uuid.UUID) (*model.UdhaarAccount, error) {
	var a model.UdhaarAccount
	err := r.pool.QueryRow(ctx, `
		SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
		FROM udhaar_accounts WHERE vendor_id = $1 AND customer_id = $2
	`, vendorID, customerID).Scan(&a.ID, &a.VendorID, &a.CustomerID, &a.Status, &a.CreditLimitPaise, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("udhaar account")
		}
		return nil, err
	}
	return &a, nil
}

// GetAccountsByVendor returns paginated udhaar accounts for a vendor (cursor = last seen ID as string).
func (r *UdhaarRepository) GetAccountsByVendor(ctx context.Context, vendorID uuid.UUID, limit int, cursor string) ([]model.UdhaarAccount, error) {
	var rows pgx.Rows
	var err error

	if cursor == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
			FROM udhaar_accounts WHERE vendor_id = $1
			ORDER BY created_at DESC, id DESC LIMIT $2
		`, vendorID, limit+1)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
			FROM udhaar_accounts WHERE vendor_id = $1 AND id < $2::uuid
			ORDER BY created_at DESC, id DESC LIMIT $3
		`, vendorID, cursor, limit+1)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

// GetAccountsByCustomer returns paginated udhaar accounts for a customer.
func (r *UdhaarRepository) GetAccountsByCustomer(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]model.UdhaarAccount, error) {
	var rows pgx.Rows
	var err error

	if cursor == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
			FROM udhaar_accounts WHERE customer_id = $1
			ORDER BY created_at DESC, id DESC LIMIT $2
		`, customerID, limit+1)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
			FROM udhaar_accounts WHERE customer_id = $1 AND id < $2::uuid
			ORDER BY created_at DESC, id DESC LIMIT $3
		`, customerID, cursor, limit+1)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func scanAccounts(rows pgx.Rows) ([]model.UdhaarAccount, error) {
	var accounts []model.UdhaarAccount
	for rows.Next() {
		var a model.UdhaarAccount
		if err := rows.Scan(&a.ID, &a.VendorID, &a.CustomerID, &a.Status, &a.CreditLimitPaise, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// GetAccountBalance computes the outstanding balance for an account.
// Balance = SUM(CREDIT amount) + SUM(ADJUSTMENT amount) - SUM(REPAYMENT amount).
// Since all amounts are > 0 and entry_type determines sign:
//
//	CREDIT and ADJUSTMENT -> add to balance (outstanding debt)
//	REPAYMENT -> subtract from balance
func (r *UdhaarRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(
			CASE entry_type
				WHEN 'CREDIT'     THEN  amount_paise
				WHEN 'ADJUSTMENT' THEN  amount_paise
				WHEN 'REPAYMENT'  THEN -amount_paise
				ELSE 0
			END
		), 0)
		FROM udhaar_entries WHERE account_id = $1
	`, accountID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

// AddEntry inserts a new udhaar entry. For REPAYMENT, checks balance > amount.
// Uses a transaction with SELECT FOR UPDATE to prevent concurrent over-repayment.
func (r *UdhaarRepository) AddEntry(ctx context.Context, accountID uuid.UUID, actorUserID uuid.UUID, input model.AddUdhaarEntryInput) (*model.UdhaarEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock the account row to prevent concurrent mutations.
	var lockedID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM udhaar_accounts WHERE id = $1 FOR UPDATE`, accountID).Scan(&lockedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("udhaar account")
		}
		return nil, err
	}

	if input.EntryType == string(model.UdhaarEntryTypeRepayment) {
		// Compute current balance within transaction.
		var balance int64
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(
				CASE entry_type
					WHEN 'CREDIT'     THEN  amount_paise
					WHEN 'ADJUSTMENT' THEN  amount_paise
					WHEN 'REPAYMENT'  THEN -amount_paise
					ELSE 0
				END
			), 0)
			FROM udhaar_entries WHERE account_id = $1
		`, accountID).Scan(&balance)
		if err != nil {
			return nil, err
		}
		if input.AmountPaise > balance {
			return nil, shared.NewUdhaarLimitExceeded(balance, input.AmountPaise)
		}
	}

	id := uuid.New()
	var e model.UdhaarEntry
	err = tx.QueryRow(ctx, `
		INSERT INTO udhaar_entries
		  (id, account_id, entry_type, amount_paise, reference_type, reference_id, idempotency_key, actor_user_id, note, created_at)
		VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, $8, NOW())
		RETURNING id, account_id, entry_type, amount_paise, reference_type, reference_id, idempotency_key, actor_user_id, note, created_at
	`, id, accountID, input.EntryType, input.AmountPaise, input.ReferenceType, input.IdempotencyKey, actorUserID, input.Note,
	).Scan(&e.ID, &e.AccountID, &e.EntryType, &e.AmountPaise, &e.ReferenceType, &e.ReferenceID, &e.IdempotencyKey, &e.ActorUserID, &e.Note, &e.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, shared.NewConflict("entry with this idempotency key already recorded for this account")
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &e, nil
}

// GetEntries returns paginated udhaar entries for an account.
func (r *UdhaarRepository) GetEntries(ctx context.Context, accountID uuid.UUID, limit int, cursor string) ([]model.UdhaarEntry, error) {
	var rows pgx.Rows
	var err error

	if cursor == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, account_id, entry_type, amount_paise, reference_type, reference_id, idempotency_key, actor_user_id, note, created_at
			FROM udhaar_entries WHERE account_id = $1
			ORDER BY created_at DESC, id DESC LIMIT $2
		`, accountID, limit+1)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, account_id, entry_type, amount_paise, reference_type, reference_id, idempotency_key, actor_user_id, note, created_at
			FROM udhaar_entries WHERE account_id = $1 AND id < $2::uuid
			ORDER BY created_at DESC, id DESC LIMIT $3
		`, accountID, cursor, limit+1)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.UdhaarEntry
	for rows.Next() {
		var e model.UdhaarEntry
		if err := rows.Scan(&e.ID, &e.AccountID, &e.EntryType, &e.AmountPaise, &e.ReferenceType, &e.ReferenceID, &e.IdempotencyKey, &e.ActorUserID, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// UpdateAccountStatus updates the status of a udhaar account.
func (r *UdhaarRepository) UpdateAccountStatus(ctx context.Context, accountID uuid.UUID, status model.UdhaarAccountStatus) (*model.UdhaarAccount, error) {
	var a model.UdhaarAccount
	err := r.pool.QueryRow(ctx, `
		UPDATE udhaar_accounts SET status = $2, updated_at = NOW() WHERE id = $1
		RETURNING id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at
	`, accountID, status,
	).Scan(&a.ID, &a.VendorID, &a.CustomerID, &a.Status, &a.CreditLimitPaise, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.NewNotFound("udhaar account")
		}
		return nil, err
	}
	return &a, nil
}

// GetVendorProfileByOwnerUserID fetches the vendor profile id for the given user.
func (r *UdhaarRepository) GetVendorProfileByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (uuid.UUID, error) {
	var vendorID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM vendor_profiles WHERE owner_user_id = $1
	`, ownerUserID).Scan(&vendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, shared.NewNotFound("vendor profile")
		}
		return uuid.Nil, err
	}
	return vendorID, nil
}
