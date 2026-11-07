package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/sync/model"
	"vendorsaathi/internal/modules/sync/repository"
	"vendorsaathi/internal/shared"
)

// SyncService owns business logic for the offline sync endpoint.
type SyncService struct {
	syncRepo *repository.SyncRepository
	pool     *pgxpool.Pool
}

// NewSyncService constructs a SyncService.
func NewSyncService(syncRepo *repository.SyncRepository, pool *pgxpool.Pool) *SyncService {
	return &SyncService{syncRepo: syncRepo, pool: pool}
}

// SubmitOperation handles an allowlisted offline sync operation.
//
//  1. Deduplicates by (user_id, client_operation_id) — returns existing if already processed.
//  2. Validates that operation_type is in the allowlist.
//  3. Creates a sync_operation record as PENDING.
//  4. Processes the operation synchronously:
//     - PRODUCT_DESCRIPTION_UPDATE: validates product ownership, updates description.
//     - STOCK_ADJUSTMENT_REQUEST: validates product ownership, adjusts on_hand_qty.
//  5. Updates status to ACCEPTED or REJECTED with result payload.
func (s *SyncService) SubmitOperation(ctx context.Context, userID uuid.UUID, input model.SubmitSyncOperationInput) (*model.SyncOperationOutput, error) {
	// 1. Deduplication
	existing, err := s.syncRepo.FindByClientID(ctx, userID, input.ClientOperationID)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("sync dedup lookup: %w", err)
	}
	if existing != nil {
		out := existing.ToOutput()
		return &out, nil
	}

	// 2. Allowlist check
	if !model.AllowedOperationTypes[input.OperationType] {
		return nil, shared.NewAppError("REJECTED", fmt.Sprintf("operation_type %q is not supported", input.OperationType), shared.ErrValidationFailed)
	}

	// 3. Create PENDING record
	op := &model.SyncOperation{
		ID:                uuid.New(),
		UserID:            userID,
		DeviceID:          input.DeviceID,
		ClientOperationID: input.ClientOperationID,
		OperationType:     input.OperationType,
		PayloadVersion:    input.PayloadVersion,
		Status:            model.SyncStatusPending,
		Payload:           input.Payload,
		CreatedAt:         time.Now().UTC(),
	}

	created, err := s.syncRepo.Create(ctx, op)
	if err != nil {
		return nil, fmt.Errorf("create sync operation: %w", err)
	}

	// 4. Process synchronously
	now := time.Now().UTC()
	var resultStatus model.SyncOperationStatus
	var resultPayload map[string]interface{}

	processErr := s.processOperation(ctx, userID, created, input.Payload)
	if processErr != nil {
		resultStatus = model.SyncStatusRejected
		resultPayload = map[string]interface{}{"error": processErr.Error()}
	} else {
		resultStatus = model.SyncStatusAccepted
		resultPayload = map[string]interface{}{"message": "operation applied successfully"}
	}

	// 5. Persist final status
	final, err := s.syncRepo.UpdateStatus(ctx, created.ID, resultStatus, resultPayload, &now)
	if err != nil {
		return nil, fmt.Errorf("update sync operation status: %w", err)
	}

	out := final.ToOutput()
	return &out, nil
}

// processOperation dispatches to the correct handler for the operation type.
func (s *SyncService) processOperation(ctx context.Context, userID uuid.UUID, op *model.SyncOperation, payload map[string]interface{}) error {
	switch op.OperationType {
	case model.OpTypeProductDescriptionUpdate:
		return s.applyProductDescriptionUpdate(ctx, userID, payload)
	case model.OpTypeStockAdjustmentRequest:
		return s.applyStockAdjustmentRequest(ctx, userID, payload)
	default:
		return fmt.Errorf("unsupported operation type: %s", op.OperationType)
	}
}

// applyProductDescriptionUpdate validates product ownership and updates its description.
// Expected payload: { "product_id": "<uuid>", "description": "<string>" }
func (s *SyncService) applyProductDescriptionUpdate(ctx context.Context, userID uuid.UUID, payload map[string]interface{}) error {
	productID, err := extractUUID(payload, "product_id")
	if err != nil {
		return err
	}

	description, ok := payload["description"].(string)
	if !ok {
		return fmt.Errorf("payload missing or invalid field: description")
	}

	// Ownership check
	var vendorID uuid.UUID
	err = s.pool.QueryRow(ctx,
		`SELECT vendor_id FROM products WHERE id = $1`, productID,
	).Scan(&vendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("product not found")
		}
		return err
	}
	if vendorID != userID {
		return fmt.Errorf("product not owned by user")
	}

	// Update description
	_, err = s.pool.Exec(ctx,
		`UPDATE products SET description = $1, updated_at = NOW() WHERE id = $2`,
		description, productID,
	)
	return err
}

// applyStockAdjustmentRequest validates product ownership and adjusts on_hand_qty.
// Expected payload: { "product_id": "<uuid>", "qty_delta": <int64>, "reason": "<string>" }
func (s *SyncService) applyStockAdjustmentRequest(ctx context.Context, userID uuid.UUID, payload map[string]interface{}) error {
	productID, err := extractUUID(payload, "product_id")
	if err != nil {
		return err
	}

	qtyDelta, err := extractInt64(payload, "qty_delta")
	if err != nil {
		return err
	}

	// Ownership check
	var vendorID uuid.UUID
	err = s.pool.QueryRow(ctx,
		`SELECT vendor_id FROM products WHERE id = $1`, productID,
	).Scan(&vendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("product not found")
		}
		return err
	}
	if vendorID != userID {
		return fmt.Errorf("product not owned by user")
	}

	// Adjust inventory in a transaction (optimistic: no reservation check for async requests)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var onHand int64
	err = tx.QueryRow(ctx,
		`SELECT on_hand_qty FROM inventory WHERE product_id = $1 FOR UPDATE`, productID,
	).Scan(&onHand)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("inventory record not found for product")
		}
		return err
	}

	newQty := onHand + qtyDelta
	if newQty < 0 {
		return fmt.Errorf("adjustment would result in negative stock: available=%d requested_delta=%d", onHand, qtyDelta)
	}

	_, err = tx.Exec(ctx,
		`UPDATE inventory SET on_hand_qty = $1, version = version+1, updated_at = NOW() WHERE product_id = $2`,
		newQty, productID,
	)
	if err != nil {
		return err
	}

	// Record a movement entry
	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, created_at)
		VALUES ($1, $2, 'ADJUSTMENT', $3, $4, $5, NOW())
	`, uuid.New(), productID, qtyDelta, payload["reason"], userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetMyOperations returns a user's paginated sync operations.
func (s *SyncService) GetMyOperations(ctx context.Context, userID uuid.UUID, deviceID string, limit int, cursor string) (*model.SyncOperationListOutput, error) {
	if limit <= 0 {
		limit = 20
	}

	ops, nextCursor, hasMore, err := s.syncRepo.GetByUser(ctx, userID, deviceID, limit, cursor)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.SyncOperationOutput, len(ops))
	for i, op := range ops {
		outputs[i] = op.ToOutput()
	}

	return &model.SyncOperationListOutput{
		Data:       outputs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// --- payload extraction helpers ---

func extractUUID(payload map[string]interface{}, key string) (uuid.UUID, error) {
	v, ok := payload[key]
	if !ok {
		return uuid.UUID{}, fmt.Errorf("payload missing field: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("payload field %s must be a string UUID", key)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("payload field %s is not a valid UUID: %w", key, err)
	}
	return id, nil
}

func extractInt64(payload map[string]interface{}, key string) (int64, error) {
	v, ok := payload[key]
	if !ok {
		return 0, fmt.Errorf("payload missing field: %s", key)
	}
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("payload field %s must be numeric", key)
	}
}
