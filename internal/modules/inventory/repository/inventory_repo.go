package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/inventory/model"
	"vendorsaathi/internal/shared"
)

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

func (r *InventoryRepository) GetInventory(ctx context.Context, productID uuid.UUID) (*model.Inventory, error) {
	var inv model.Inventory
	err := r.pool.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepository) GetInventoryForUpdate(ctx context.Context, productID uuid.UUID) (*model.Inventory, error) {
	var inv model.Inventory
	err := r.pool.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1 FOR UPDATE
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepository) CreateInventory(ctx context.Context, productID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO inventory (product_id, on_hand_qty, reserved_qty, version, updated_at)
		VALUES ($1, 0, 0, 0, NOW())
		ON CONFLICT (product_id) DO NOTHING
	`, productID)
	return err
}

func (r *InventoryRepository) AdjustStock(ctx context.Context, productID uuid.UUID, qtyDelta int64, reason *string, actorUserID *uuid.UUID, idempotencyKey *string) (*model.Inventory, *model.InventoryMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != nil {
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM inventory_movements WHERE idempotency_key = $1)
		`, *idempotencyKey).Scan(&exists)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return nil, nil, shared.NewIdempotencyConflict()
		}
	}

	var inv model.Inventory
	err = tx.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1 FOR UPDATE
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.ErrNotFound
		}
		return nil, nil, err
	}

	newOnHand := inv.OnHandQty + qtyDelta
	if newOnHand < 0 {
		return nil, nil, shared.NewInsufficientStock(inv.OnHandQty, -qtyDelta)
	}
	if newOnHand < inv.ReservedQty {
		return nil, nil, shared.NewInsufficientStock(inv.AvailableQty(), -qtyDelta)
	}

	newVersion := inv.Version + 1
	err = tx.QueryRow(ctx, `
		UPDATE inventory
		SET on_hand_qty = $1, version = $2, updated_at = NOW()
		WHERE product_id = $3 AND version = $4
		RETURNING product_id, on_hand_qty, reserved_qty, version, updated_at
	`, newOnHand, newVersion, productID, inv.Version).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.NewOptimisticLock()
		}
		return nil, nil, err
	}

	movement := &model.InventoryMovement{
		ID:             uuid.New(),
		ProductID:      productID,
		MovementType:   model.MovementTypeAdjustment,
		QtyDelta:       qtyDelta,
		Reason:         reason,
		ActorUserID:    actorUserID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      shared.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, movement.ID, movement.ProductID, movement.MovementType, movement.QtyDelta, movement.Reason, movement.ActorUserID, movement.IdempotencyKey, movement.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, shared.NewIdempotencyConflict()
		}
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &inv, movement, nil
}

func (r *InventoryRepository) ReserveStock(ctx context.Context, productID uuid.UUID, qty int64, idempotencyKey *string, referenceType *string, referenceID *uuid.UUID, actorUserID *uuid.UUID) (*model.Inventory, *model.InventoryMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != nil {
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM inventory_movements WHERE idempotency_key = $1)
		`, *idempotencyKey).Scan(&exists)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return nil, nil, shared.NewIdempotencyConflict()
		}
	}

	var inv model.Inventory
	err = tx.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1 FOR UPDATE
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.ErrNotFound
		}
		return nil, nil, err
	}

	available := inv.OnHandQty - inv.ReservedQty
	if qty > available {
		return nil, nil, shared.NewInsufficientStock(available, qty)
	}

	newReserved := inv.ReservedQty + qty
	newVersion := inv.Version + 1
	err = tx.QueryRow(ctx, `
		UPDATE inventory
		SET reserved_qty = $1, version = $2, updated_at = NOW()
		WHERE product_id = $3 AND version = $4
		RETURNING product_id, on_hand_qty, reserved_qty, version, updated_at
	`, newReserved, newVersion, productID, inv.Version).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.NewOptimisticLock()
		}
		return nil, nil, err
	}

	refTypePtr := referenceType
	if refTypePtr == nil {
		refTypeStr := "RESERVATION"
		refTypePtr = &refTypeStr
	}

	movement := &model.InventoryMovement{
		ID:             uuid.New(),
		ProductID:      productID,
		MovementType:   model.MovementTypeReservation,
		QtyDelta:       -qty,
		Reason:         nil,
		ActorUserID:    actorUserID,
		ReferenceType:  refTypePtr,
		ReferenceID:    referenceID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      shared.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, movement.ID, movement.ProductID, movement.MovementType, movement.QtyDelta, movement.Reason, movement.ActorUserID, movement.ReferenceType, movement.ReferenceID, movement.IdempotencyKey, movement.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, shared.NewIdempotencyConflict()
		}
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &inv, movement, nil
}

func (r *InventoryRepository) ReleaseStock(ctx context.Context, productID uuid.UUID, qty int64, idempotencyKey *string, referenceType *string, referenceID *uuid.UUID, actorUserID *uuid.UUID) (*model.Inventory, *model.InventoryMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != nil {
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM inventory_movements WHERE idempotency_key = $1)
		`, *idempotencyKey).Scan(&exists)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return nil, nil, shared.NewIdempotencyConflict()
		}
	}

	var inv model.Inventory
	err = tx.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1 FOR UPDATE
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.ErrNotFound
		}
		return nil, nil, err
	}

	if qty > inv.ReservedQty {
		return nil, nil, shared.NewInsufficientStock(inv.ReservedQty, qty)
	}

	newReserved := inv.ReservedQty - qty
	newVersion := inv.Version + 1
	err = tx.QueryRow(ctx, `
		UPDATE inventory
		SET reserved_qty = $1, version = $2, updated_at = NOW()
		WHERE product_id = $3 AND version = $4
		RETURNING product_id, on_hand_qty, reserved_qty, version, updated_at
	`, newReserved, newVersion, productID, inv.Version).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.NewOptimisticLock()
		}
		return nil, nil, err
	}

	refTypePtr := referenceType
	if refTypePtr == nil {
		refTypeStr := "RELEASE"
		refTypePtr = &refTypeStr
	}

	movement := &model.InventoryMovement{
		ID:             uuid.New(),
		ProductID:      productID,
		MovementType:   model.MovementTypeRelease,
		QtyDelta:       qty,
		Reason:         nil,
		ActorUserID:    actorUserID,
		ReferenceType:  refTypePtr,
		ReferenceID:    referenceID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      shared.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, movement.ID, movement.ProductID, movement.MovementType, movement.QtyDelta, movement.Reason, movement.ActorUserID, movement.ReferenceType, movement.ReferenceID, movement.IdempotencyKey, movement.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, shared.NewIdempotencyConflict()
		}
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &inv, movement, nil
}

func (r *InventoryRepository) FulfilStock(ctx context.Context, productID uuid.UUID, qty int64, idempotencyKey *string, referenceType *string, referenceID *uuid.UUID, actorUserID *uuid.UUID) (*model.Inventory, *model.InventoryMovement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != nil {
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM inventory_movements WHERE idempotency_key = $1)
		`, *idempotencyKey).Scan(&exists)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			return nil, nil, shared.NewIdempotencyConflict()
		}
	}

	var inv model.Inventory
	err = tx.QueryRow(ctx, `
		SELECT product_id, on_hand_qty, reserved_qty, version, updated_at
		FROM inventory WHERE product_id = $1 FOR UPDATE
	`, productID).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.ErrNotFound
		}
		return nil, nil, err
	}

	if qty > inv.ReservedQty {
		return nil, nil, shared.NewInsufficientStock(inv.ReservedQty, qty)
	}

	newOnHand := inv.OnHandQty - qty
	newReserved := inv.ReservedQty - qty
	newVersion := inv.Version + 1
	err = tx.QueryRow(ctx, `
		UPDATE inventory
		SET on_hand_qty = $1, reserved_qty = $2, version = $3, updated_at = NOW()
		WHERE product_id = $4 AND version = $5
		RETURNING product_id, on_hand_qty, reserved_qty, version, updated_at
	`, newOnHand, newReserved, newVersion, productID, inv.Version).Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, shared.NewOptimisticLock()
		}
		return nil, nil, err
	}

	refTypePtr := referenceType
	if refTypePtr == nil {
		refTypeStr := "FULFILMENT"
		refTypePtr = &refTypeStr
	}

	movement := &model.InventoryMovement{
		ID:             uuid.New(),
		ProductID:      productID,
		MovementType:   model.MovementTypeFulfilment,
		QtyDelta:       -qty,
		Reason:         nil,
		ActorUserID:    actorUserID,
		ReferenceType:  refTypePtr,
		ReferenceID:    referenceID,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      shared.Now(),
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, movement.ID, movement.ProductID, movement.MovementType, movement.QtyDelta, movement.Reason, movement.ActorUserID, movement.ReferenceType, movement.ReferenceID, movement.IdempotencyKey, movement.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, shared.NewIdempotencyConflict()
		}
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &inv, movement, nil
}

func (r *InventoryRepository) GetMovements(ctx context.Context, productID uuid.UUID, limit int, cursor string) ([]model.InventoryMovement, string, bool, error) {
	query := `
		SELECT id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, idempotency_key, created_at
		FROM inventory_movements WHERE product_id = $1
	`
	args := []interface{}{productID}
	argIdx := 2

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(argIdx+'0'))
	args = append(args, limit+1)
	argIdx++

	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, "", false, err
		}
		query += ` AND created_at < $` + string(rune(argIdx+'0'))
		args = append(args, cursorTime)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var movements []model.InventoryMovement
	for rows.Next() {
		var m model.InventoryMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.MovementType, &m.QtyDelta, &m.Reason, &m.ActorUserID, &m.ReferenceType, &m.ReferenceID, &m.IdempotencyKey, &m.CreatedAt); err != nil {
			return nil, "", false, err
		}
		movements = append(movements, m)
	}

	hasMore := len(movements) > limit
	if hasMore {
		movements = movements[:limit]
	}

	var nextCursor string
	if hasMore && len(movements) > 0 {
		nextCursor = movements[len(movements)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return movements, nextCursor, hasMore, rows.Err()
}

func (r *InventoryRepository) GetInventoryByVendor(ctx context.Context, vendorID uuid.UUID) ([]model.Inventory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.product_id, i.on_hand_qty, i.reserved_qty, i.version, i.updated_at
		FROM inventory i
		JOIN products p ON i.product_id = p.id
		WHERE (p.vendor_id = $1 OR p.vendor_id IN (SELECT id FROM vendor_profiles WHERE owner_user_id = $1))
		ORDER BY p.name
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []model.Inventory
	for rows.Next() {
		var inv model.Inventory
		if err := rows.Scan(&inv.ProductID, &inv.OnHandQty, &inv.ReservedQty, &inv.Version, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}
	return inventories, rows.Err()
}
