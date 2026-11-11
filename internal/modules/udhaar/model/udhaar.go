package model

import (
	"time"

	"github.com/google/uuid"
)

type UdhaarAccountStatus string

const (
	UdhaarAccountStatusActive    UdhaarAccountStatus = "ACTIVE"
	UdhaarAccountStatusClosed    UdhaarAccountStatus = "CLOSED"
	UdhaarAccountStatusSuspended UdhaarAccountStatus = "SUSPENDED"
)

type UdhaarEntryType string

const (
	UdhaarEntryTypeCredit     UdhaarEntryType = "CREDIT"
	UdhaarEntryTypeRepayment  UdhaarEntryType = "REPAYMENT"
	UdhaarEntryTypeAdjustment UdhaarEntryType = "ADJUSTMENT"
)

type UdhaarAccount struct {
	ID               uuid.UUID           `db:"id"`
	VendorID         uuid.UUID           `db:"vendor_id"`
	CustomerID       uuid.UUID           `db:"customer_id"`
	CustomerName     *string             `db:"customer_name"`
	CustomerPhone    *string             `db:"customer_phone"`
	Status           UdhaarAccountStatus `db:"status"`
	CreditLimitPaise *int64              `db:"credit_limit_paise"`
	CreatedAt        time.Time           `db:"created_at"`
	UpdatedAt        time.Time           `db:"updated_at"`
}

type UdhaarEntry struct {
	ID             uuid.UUID       `db:"id"`
	AccountID      uuid.UUID       `db:"account_id"`
	EntryType      UdhaarEntryType `db:"entry_type"`
	AmountPaise    int64           `db:"amount_paise"`
	ReferenceType  *string         `db:"reference_type"`
	ReferenceID    *uuid.UUID      `db:"reference_id"`
	IdempotencyKey *string         `db:"idempotency_key"`
	ActorUserID    uuid.UUID       `db:"actor_user_id"`
	Note           *string         `db:"note"`
	CreatedAt      time.Time       `db:"created_at"`
}

// --- Input types ---

type CreateUdhaarAccountInput struct {
	CustomerID       uuid.UUID `json:"customer_id" binding:"required"`
	CreditLimitPaise *int64    `json:"credit_limit_paise,omitempty"`
}

type AddUdhaarEntryInput struct {
	EntryType      string  `json:"entry_type" binding:"required,oneof=CREDIT REPAYMENT ADJUSTMENT"`
	AmountPaise    int64   `json:"amount_paise" binding:"required,min=1"`
	ReferenceType  *string `json:"reference_type,omitempty"`
	Note           *string `json:"note,omitempty"`
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

type UpdateUdhaarAccountStatusInput struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE CLOSED SUSPENDED"`
}

// --- Output types ---

type UdhaarAccountOutput struct {
	ID               uuid.UUID           `json:"id"`
	VendorID         uuid.UUID           `json:"vendor_id"`
	CustomerID       uuid.UUID           `json:"customer_id"`
	CustomerName     *string             `json:"customer_name,omitempty"`
	CustomerPhone    *string             `json:"customer_phone,omitempty"`
	Status           UdhaarAccountStatus `json:"status"`
	CreditLimitPaise *int64              `json:"credit_limit_paise,omitempty"`
	BalancePaise     int64               `json:"balance_paise"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type UdhaarEntryOutput struct {
	ID             uuid.UUID       `json:"id"`
	AccountID      uuid.UUID       `json:"account_id"`
	EntryType      UdhaarEntryType `json:"entry_type"`
	AmountPaise    int64           `json:"amount_paise"`
	ReferenceType  *string         `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID      `json:"reference_id,omitempty"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	ActorUserID    uuid.UUID       `json:"actor_user_id"`
	Note           *string         `json:"note,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (a *UdhaarAccount) ToOutput(balancePaise int64) UdhaarAccountOutput {
	return UdhaarAccountOutput{
		ID:               a.ID,
		VendorID:         a.VendorID,
		CustomerID:       a.CustomerID,
		CustomerName:     a.CustomerName,
		CustomerPhone:    a.CustomerPhone,
		Status:           a.Status,
		CreditLimitPaise: a.CreditLimitPaise,
		BalancePaise:     balancePaise,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func (e *UdhaarEntry) ToOutput() UdhaarEntryOutput {
	return UdhaarEntryOutput{
		ID:             e.ID,
		AccountID:      e.AccountID,
		EntryType:      e.EntryType,
		AmountPaise:    e.AmountPaise,
		ReferenceType:  e.ReferenceType,
		ReferenceID:    e.ReferenceID,
		IdempotencyKey: e.IdempotencyKey,
		ActorUserID:    e.ActorUserID,
		Note:           e.Note,
		CreatedAt:      e.CreatedAt,
	}
}
