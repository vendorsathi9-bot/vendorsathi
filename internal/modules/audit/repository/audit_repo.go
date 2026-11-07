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
	"vendorsaathi/internal/modules/audit/model"
	"vendorsaathi/internal/shared"
)

// AuditRepository handles parameterised SQL for the audit_logs table.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository constructs an AuditRepository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// auditCursor is used for opaque cursor encoding/decoding.
type auditCursor struct {
	T time.Time `json:"t"`
	I uuid.UUID `json:"i"`
}

func encodeAuditCursor(t time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(auditCursor{T: t, I: id})
	return string(b)
}

func decodeAuditCursor(cursor string) (time.Time, uuid.UUID, error) {
	var c auditCursor
	if err := json.Unmarshal([]byte(cursor), &c); err != nil {
		return time.Time{}, uuid.UUID{}, err
	}
	return c.T, c.I, nil
}

// Create inserts a new audit log entry.
func (r *AuditRepository) Create(ctx context.Context, input model.CreateAuditLogInput) (*model.AuditLog, error) {
	id := uuid.New()
	now := time.Now().UTC()

	var metaBytes []byte
	if input.Metadata != nil {
		var err error
		metaBytes, err = json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshal metadata: %w", err)
		}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs
		  (id, actor_user_id, action, resource_type, resource_id, reason, request_id, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, id, input.ActorUserID, input.Action, input.ResourceType, input.ResourceID,
		input.Reason, input.RequestID, metaBytes, now)
	if err != nil {
		return nil, err
	}

	return &model.AuditLog{
		ID:           id,
		ActorUserID:  input.ActorUserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Reason:       input.Reason,
		RequestID:    input.RequestID,
		Metadata:     input.Metadata,
		CreatedAt:    now,
	}, nil
}

// AuditListFilters holds optional filter parameters for List.
type AuditListFilters struct {
	ActorUserID  *uuid.UUID
	ResourceType *string
	ResourceID   *uuid.UUID
}

// List returns a cursor-paginated list of audit logs (newest first).
func (r *AuditRepository) List(ctx context.Context, filters AuditListFilters, limit int, cursor string) ([]model.AuditLog, string, bool, error) {
	args := []interface{}{}
	argIdx := 1
	conditions := []string{}

	if cursor != "" {
		cursorTime, cursorID, err := decodeAuditCursor(cursor)
		if err == nil {
			conditions = append(conditions,
				fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIdx, argIdx+1))
			args = append(args, cursorTime, cursorID)
			argIdx += 2
		}
	}

	if filters.ActorUserID != nil {
		conditions = append(conditions, fmt.Sprintf("actor_user_id = $%d", argIdx))
		args = append(args, *filters.ActorUserID)
		argIdx++
	}
	if filters.ResourceType != nil {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argIdx))
		args = append(args, *filters.ResourceType)
		argIdx++
	}
	if filters.ResourceID != nil {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIdx))
		args = append(args, *filters.ResourceID)
		argIdx++
	}

	where := ""
	for i, cond := range conditions {
		if i == 0 {
			where = " WHERE " + cond
		} else {
			where += " AND " + cond
		}
	}

	query := fmt.Sprintf(`
		SELECT id, actor_user_id, action, resource_type, resource_id, reason, request_id, metadata, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var al model.AuditLog
		var metaBytes []byte
		if err := rows.Scan(
			&al.ID, &al.ActorUserID, &al.Action, &al.ResourceType, &al.ResourceID,
			&al.Reason, &al.RequestID, &metaBytes, &al.CreatedAt,
		); err != nil {
			return nil, "", false, err
		}
		if metaBytes != nil {
			if err := json.Unmarshal(metaBytes, &al.Metadata); err != nil {
				return nil, "", false, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		logs = append(logs, al)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, err
	}

	hasMore := len(logs) > limit
	if hasMore {
		logs = logs[:limit]
	}

	var nextCursor string
	if hasMore && len(logs) > 0 {
		last := logs[len(logs)-1]
		nextCursor = encodeAuditCursor(last.CreatedAt, last.ID)
	}

	return logs, nextCursor, hasMore, nil
}

// GetByID fetches a single audit log by its primary key.
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var al model.AuditLog
	var metaBytes []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, actor_user_id, action, resource_type, resource_id, reason, request_id, metadata, created_at
		FROM audit_logs WHERE id = $1
	`, id).Scan(
		&al.ID, &al.ActorUserID, &al.Action, &al.ResourceType, &al.ResourceID,
		&al.Reason, &al.RequestID, &metaBytes, &al.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	if metaBytes != nil {
		if err := json.Unmarshal(metaBytes, &al.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}
	return &al, nil
}
