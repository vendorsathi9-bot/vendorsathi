package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/udhaar/model"
	"vendorsaathi/internal/modules/udhaar/repository"
	"vendorsaathi/internal/shared"
)

type UdhaarService struct {
	repo *repository.UdhaarRepository
}

func NewUdhaarService(repo *repository.UdhaarRepository) *UdhaarService {
	return &UdhaarService{repo: repo}
}

// CreateAccount allows a vendor to open an Udhaar account for a customer.
func (s *UdhaarService) CreateAccount(ctx context.Context, vendorUserID uuid.UUID, input model.CreateUdhaarAccountInput) (*model.UdhaarAccountOutput, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, vendorUserID)
	if err != nil {
		return nil, shared.NewForbidden("user does not have a vendor profile")
	}

	acc, err := s.repo.CreateAccount(ctx, vendorID, input.CustomerID, input.CreditLimitPaise)
	if err != nil {
		return nil, err
	}

	out := acc.ToOutput(0)
	return &out, nil
}

// GetAccount retrieves account details and current balance.
func (s *UdhaarService) GetAccount(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID) (*model.UdhaarAccountOutput, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Verify actor is either the customer or the vendor owner
	isCustomer := acc.CustomerID == actorUserID
	var isVendor bool
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err == nil && vendorID == acc.VendorID {
		isVendor = true
	}

	if !isCustomer && !isVendor {
		return nil, shared.NewForbidden("access to udhaar account denied")
	}

	balance, err := s.repo.GetAccountBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}

	out := acc.ToOutput(balance)
	return &out, nil
}

// GetAccountsByVendor lists accounts belonging to the vendor.
func (s *UdhaarService) GetAccountsByVendor(ctx context.Context, vendorUserID uuid.UUID, limit int, cursor string) ([]model.UdhaarAccountOutput, string, bool, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, vendorUserID)
	if err != nil {
		return nil, "", false, shared.NewForbidden("user does not have a vendor profile")
	}

	if limit <= 0 {
		limit = 20
	}
	accounts, err := s.repo.GetAccountsByVendor(ctx, vendorID, limit, cursor)
	if err != nil {
		return nil, "", false, err
	}

	hasMore := len(accounts) > limit
	if hasMore {
		accounts = accounts[:limit]
	}

	var nextCursor string
	if len(accounts) > 0 {
		nextCursor = accounts[len(accounts)-1].ID.String()
	}

	outputs := make([]model.UdhaarAccountOutput, 0, len(accounts))
	for _, a := range accounts {
		balance, _ := s.repo.GetAccountBalance(ctx, a.ID)
		outputs = append(outputs, a.ToOutput(balance))
	}

	return outputs, nextCursor, hasMore, nil
}

// GetAccountsByCustomer lists udhaar accounts for the authenticated customer.
func (s *UdhaarService) GetAccountsByCustomer(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]model.UdhaarAccountOutput, string, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	accounts, err := s.repo.GetAccountsByCustomer(ctx, customerID, limit, cursor)
	if err != nil {
		return nil, "", false, err
	}

	hasMore := len(accounts) > limit
	if hasMore {
		accounts = accounts[:limit]
	}

	var nextCursor string
	if len(accounts) > 0 {
		nextCursor = accounts[len(accounts)-1].ID.String()
	}

	outputs := make([]model.UdhaarAccountOutput, 0, len(accounts))
	for _, a := range accounts {
		balance, _ := s.repo.GetAccountBalance(ctx, a.ID)
		outputs = append(outputs, a.ToOutput(balance))
	}

	return outputs, nextCursor, hasMore, nil
}

// AddEntry records a CREDIT, REPAYMENT, or ADJUSTMENT on the account.
func (s *UdhaarService) AddEntry(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, input model.AddUdhaarEntryInput) (*model.UdhaarEntryOutput, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Only vendor owner can add entries for MVP
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err != nil || vendorID != acc.VendorID {
		return nil, shared.NewForbidden("only the vendor can record udhaar entries")
	}

	if acc.Status != model.UdhaarAccountStatusActive {
		return nil, shared.NewConflict("cannot record entries on inactive udhaar account")
	}

	entry, err := s.repo.AddEntry(ctx, accountID, actorUserID, input)
	if err != nil {
		return nil, err
	}

	out := entry.ToOutput()
	return &out, nil
}

// GetEntries returns paginated ledger entries.
func (s *UdhaarService) GetEntries(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, limit int, cursor string) ([]model.UdhaarEntryOutput, string, bool, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, "", false, err
	}

	isCustomer := acc.CustomerID == actorUserID
	var isVendor bool
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, actorUserID)
	if err == nil && vendorID == acc.VendorID {
		isVendor = true
	}

	if !isCustomer && !isVendor {
		return nil, "", false, shared.NewForbidden("access to udhaar entries denied")
	}

	if limit <= 0 {
		limit = 20
	}
	entries, err := s.repo.GetEntries(ctx, accountID, limit, cursor)
	if err != nil {
		return nil, "", false, err
	}

	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}

	var nextCursor string
	if len(entries) > 0 {
		nextCursor = entries[len(entries)-1].ID.String()
	}

	outputs := make([]model.UdhaarEntryOutput, 0, len(entries))
	for _, e := range entries {
		outputs = append(outputs, e.ToOutput())
	}

	return outputs, nextCursor, hasMore, nil
}

// UpdateAccountStatus updates the status of an Udhaar account (ACTIVE, CLOSED, SUSPENDED).
func (s *UdhaarService) UpdateAccountStatus(ctx context.Context, vendorUserID uuid.UUID, accountID uuid.UUID, status model.UdhaarAccountStatus) (*model.UdhaarAccountOutput, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, vendorUserID)
	if err != nil || vendorID != acc.VendorID {
		return nil, shared.NewForbidden("only the vendor can update account status")
	}

	updated, err := s.repo.UpdateAccountStatus(ctx, accountID, status)
	if err != nil {
		return nil, err
	}

	balance, _ := s.repo.GetAccountBalance(ctx, accountID)
	out := updated.ToOutput(balance)
	return &out, nil
}
