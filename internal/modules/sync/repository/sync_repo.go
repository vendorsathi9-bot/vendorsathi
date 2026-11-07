package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/sync/model"
	"vendorsaathi/internal/shared"
)

// SyncRepository handles SQL for the sync_operations table.
type SyncRepository struct {
	pool *pgxpool.Pool
}

// NewSyncRepository constructs a SyncRepository.
func NewSyncRepository(pool *pgxpool.Pool) *SyncRepository {
	return &SyncRepository{pool: pool}
}

// --- cursor helpers ---

type syncCursor struct {
	T time.Time `json:"t"`
	I uuid.UUID `json:"i"`
}

func encodeSyncCursor(t time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(syncCursor{T: t, I: id})
	return string(b)
}

func decodeSyncCursor(cursor string) (time.Time, uuid.UUID, error) {
	var c syncCursor
	if err := json.Unmarshal([]byte(cursor), &c); err != nil {
		return time.Time{}, uuid.UUID{}, err
	}
	return c.T, c.I, nil
}

// scanSyncOp scans a row into a SyncOperation value.
func scanSyncOp(row pgx.Row) (*model.SyncOperation, error) {
	var op model.SyncOperation
	var payloadBytes, resultBytes []byte
	err := row.Scan(
		&op.ID, &op.UserID, &op.DeviceID, &op.ClientOperationID,
		&op.OperationType, &op.PayloadVersion, &op.Status,
		&payloadBytes, &resultBytes, &op.CreatedAt, &op.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	if payloadBytes != nil {
		_ = json.Unmarshal(payloadBytes, &op.Payload)
	}
	if resultBytes != nil {
		_ = json.Unmarshal(resultBytes, &op.Result)
	}
	return &op, nil
}

// Create inserts a new sync operation record.
func (r *SyncRepository) Create(ctx context.Context, op *model.SyncOperation) (*model.SyncOperation, error) {
	payloadBytes, err := json.Marshal(op.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO sync_operations
		  (id, user_id, device_id, client_operation_id, operation_type, payload_version, status, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, user_id, device_id, client_operation_id, operation_type, payload_version, status,
		          payload, result, created_at, processed_at
	`, op.ID, op.UserID, op.DeviceID, op.ClientOperationID, op.OperationType,
		op.PayloadVersion, op.Status, payloadBytes, op.CreatedAt)

	return scanSyncOp(row)
}

// FindByClientID looks up a sync operation by (user_id, client_operation_id).
func (r *SyncRepository) FindByClientID(ctx context.Context, userID, clientOpID uuid.UUID) (*model.SyncOperation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, device_id, client_operation_id, operation_type, payload_version, status,
		       payload, result, created_at, processed_at
		FROM sync_operations
		WHERE user_id = $1 AND client_operation_id = $2
	`, userID, clientOpID)
	return scanSyncOp(row)
}

// UpdateStatus updates the status, result, and processed_at of a sync operation.
func (r *SyncRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.SyncOperationStatus, result map[string]interface{}, processedAt *time.Time) (*model.SyncOperation, error) {
	var resultBytes []byte
	if result != nil {
		var err error
		resultBytes, err = json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("marshal result: %w", err)
		}
	}

	row := r.pool.QueryRow(ctx, `
		UPDATE sync_operations
		SET status = $2, result = $3, processed_at = $4
		WHERE id = $1
		RETURNING id, user_id, device_id, client_operation_id, operation_type, payload_version, status,
		          payload, result, created_at, processed_at
	`, id, status, resultBytes, processedAt)

	return scanSyncOp(row)
}

// GetByUser returns cursor-paginated sync operations for a user, optionally filtered by deviceID.
func (r *SyncRepository) GetByUser(ctx context.Context, userID uuid.UUID, deviceID string, limit int, cursor string) ([]model.SyncOperation, string, bool, error) {
	args := []interface{}{userID}
	argIdx := 2
	where := "user_id = $1"

	if deviceID != "" {
		where += fmt.Sprintf(" AND device_id = $%d", argIdx)
		args = append(args, deviceID)
		argIdx++
	}

	if cursor != "" {
		ct, cid, err := decodeSyncCursor(cursor)
		if err == nil {
			where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", argIdx, argIdx+1)
			args = append(args, ct, cid)
			argIdx += 2
		}
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, device_id, client_operation_id, operation_type, payload_version, status,
		       payload, result, created_at, processed_at
		FROM sync_operations
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var ops []model.SyncOperation
	for rows.Next() {
		var op model.SyncOperation
		var payloadBytes, resultBytes []byte
		if err := rows.Scan(
			&op.ID, &op.UserID, &op.DeviceID, &op.ClientOperationID,
			&op.OperationType, &op.PayloadVersion, &op.Status,
			&payloadBytes, &resultBytes, &op.CreatedAt, &op.ProcessedAt,
		); err != nil {
			return nil, "", false, err
		}
		if payloadBytes != nil {
			_ = json.Unmarshal(payloadBytes, &op.Payload)
		}
		if resultBytes != nil {
			_ = json.Unmarshal(resultBytes, &op.Result)
		}
		ops = append(ops, op)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, err
	}

	hasMore := len(ops) > limit
	if hasMore {
		ops = ops[:limit]
	}

	var nextCursor string
	if hasMore && len(ops) > 0 {
		last := ops[len(ops)-1]
		nextCursor = encodeSyncCursor(last.CreatedAt, last.ID)
	}

	return ops, nextCursor, hasMore, nil
}
