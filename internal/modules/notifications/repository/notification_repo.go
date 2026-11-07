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
	"vendorsaathi/internal/modules/notifications/model"
	"vendorsaathi/internal/shared"
)

// NotificationRepository handles SQL for notifications and device_tokens.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository constructs a NotificationRepository.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// --- cursor helpers ---

type notifCursor struct {
	T time.Time `json:"t"`
	I uuid.UUID `json:"i"`
}

func encodeNotifCursor(t time.Time, id uuid.UUID) string {
	b, _ := json.Marshal(notifCursor{T: t, I: id})
	return string(b)
}

func decodeNotifCursor(cursor string) (time.Time, uuid.UUID, error) {
	var c notifCursor
	if err := json.Unmarshal([]byte(cursor), &c); err != nil {
		return time.Time{}, uuid.UUID{}, err
	}
	return c.T, c.I, nil
}

// --- Notifications ---

// Create inserts a new notification record.
func (r *NotificationRepository) Create(ctx context.Context, userID uuid.UUID, eventType, title, body string, payload map[string]interface{}) (*model.Notification, error) {
	id := uuid.New()
	now := time.Now().UTC()

	var payloadBytes []byte
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (id, user_id, event_type, title, body, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, userID, eventType, title, body, payloadBytes, now)
	if err != nil {
		return nil, err
	}

	return &model.Notification{
		ID:        id,
		UserID:    userID,
		EventType: eventType,
		Title:     title,
		Body:      body,
		Payload:   payload,
		CreatedAt: now,
	}, nil
}

// GetByUser returns a cursor-paginated list of notifications for a user (newest first).
func (r *NotificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]model.Notification, string, bool, error) {
	args := []interface{}{userID}
	argIdx := 2
	where := "user_id = $1"

	if cursor != "" {
		ct, cid, err := decodeNotifCursor(cursor)
		if err == nil {
			where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", argIdx, argIdx+1)
			args = append(args, ct, cid)
			argIdx += 2
		}
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, event_type, title, body, payload, read_at, created_at
		FROM notifications
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

	var notifs []model.Notification
	for rows.Next() {
		var n model.Notification
		var payloadBytes []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.EventType, &n.Title, &n.Body, &payloadBytes, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, "", false, err
		}
		if payloadBytes != nil {
			_ = json.Unmarshal(payloadBytes, &n.Payload)
		}
		notifs = append(notifs, n)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, err
	}

	hasMore := len(notifs) > limit
	if hasMore {
		notifs = notifs[:limit]
	}

	var nextCursor string
	if hasMore && len(notifs) > 0 {
		last := notifs[len(notifs)-1]
		nextCursor = encodeNotifCursor(last.CreatedAt, last.ID)
	}

	return notifs, nextCursor, hasMore, nil
}

// MarkRead marks a notification as read; enforces ownership.
func (r *NotificationRepository) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE notifications
		SET read_at = $1
		WHERE id = $2 AND user_id = $3 AND read_at IS NULL
	`, now, notificationID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Either not found or already read – treat both as not found.
		return shared.ErrNotFound
	}
	return nil
}

// MarkAllRead marks every unread notification for a user as read.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications
		SET read_at = $1
		WHERE user_id = $2 AND read_at IS NULL
	`, now, userID)
	return err
}

// UnreadCount returns the number of unread notifications for a user.
func (r *NotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
	`, userID).Scan(&count)
	return count, err
}

// --- Device Tokens ---

// RegisterDeviceToken upserts a device token for a user.
func (r *NotificationRepository) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token string, platform model.Platform) (*model.DeviceToken, error) {
	id := uuid.New()
	now := time.Now().UTC()

	var dt model.DeviceToken
	err := r.pool.QueryRow(ctx, `
		INSERT INTO device_tokens (id, user_id, token, platform, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (token) DO UPDATE
		  SET user_id = EXCLUDED.user_id,
		      platform = EXCLUDED.platform,
		      last_seen_at = EXCLUDED.last_seen_at,
		      revoked_at = NULL
		RETURNING id, user_id, token, platform, last_seen_at, revoked_at, created_at
	`, id, userID, token, platform, now).Scan(
		&dt.ID, &dt.UserID, &dt.Token, &dt.Platform,
		&dt.LastSeenAt, &dt.RevokedAt, &dt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

// RevokeDeviceToken sets revoked_at for a specific token owned by userID.
func (r *NotificationRepository) RevokeDeviceToken(ctx context.Context, userID uuid.UUID, token string) error {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE device_tokens
		SET revoked_at = $1
		WHERE token = $2 AND user_id = $3 AND revoked_at IS NULL
	`, now, token, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetActiveDeviceTokens returns all non-revoked tokens for a user.
func (r *NotificationRepository) GetActiveDeviceTokens(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, token, platform, last_seen_at, revoked_at, created_at
		FROM device_tokens
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY last_seen_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []model.DeviceToken
	for rows.Next() {
		var dt model.DeviceToken
		if err := rows.Scan(&dt.ID, &dt.UserID, &dt.Token, &dt.Platform, &dt.LastSeenAt, &dt.RevokedAt, &dt.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, dt)
	}
	return tokens, rows.Err()
}

// RevokeInvalidToken revokes a token by its string value (called on FCM invalid-token response).
func (r *NotificationRepository) RevokeInvalidToken(ctx context.Context, token string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE device_tokens SET revoked_at = $1 WHERE token = $2
	`, now, token)
	return err
}

// GetNotificationByID fetches a single notification (used for ownership checks).
func (r *NotificationRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	var n model.Notification
	var payloadBytes []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, event_type, title, body, payload, read_at, created_at
		FROM notifications WHERE id = $1
	`, id).Scan(&n.ID, &n.UserID, &n.EventType, &n.Title, &n.Body, &payloadBytes, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	if payloadBytes != nil {
		_ = json.Unmarshal(payloadBytes, &n.Payload)
	}
	return &n, nil
}
