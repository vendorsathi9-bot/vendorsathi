package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/notifications/model"
	"vendorsaathi/internal/modules/notifications/repository"
	"vendorsaathi/internal/platform/logger"
)

// NotificationService owns notification business logic.
type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

// NewNotificationService constructs a NotificationService.
func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

// NotifyUser creates an in-app notification record for a user.
// Push delivery (FCM) is a no-op stub for MVP local development.
func (s *NotificationService) NotifyUser(ctx context.Context, userID uuid.UUID, eventType, title, body string, payload map[string]interface{}) error {
	_, err := s.notifRepo.Create(ctx, userID, eventType, title, body, payload)
	if err != nil {
		return err
	}
	// TODO(mvp): send FCM push notification to active device tokens.
	// Tokens can be fetched via s.notifRepo.GetActiveDeviceTokens(ctx, userID).
	logger.Logger.Debug().
		Str("user_id", userID.String()).
		Str("event_type", eventType).
		Msg("notification: in-app record created (FCM delivery is a no-op in MVP)")
	return nil
}

// GetMyNotifications returns paginated notifications for a user.
func (s *NotificationService) GetMyNotifications(ctx context.Context, userID uuid.UUID, limit int, cursor string) (*model.NotificationListOutput, error) {
	if limit <= 0 {
		limit = 20
	}

	notifs, nextCursor, hasMore, err := s.notifRepo.GetByUser(ctx, userID, limit, cursor)
	if err != nil {
		return nil, err
	}

	unread, err := s.notifRepo.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.NotificationOutput, len(notifs))
	for i, n := range notifs {
		outputs[i] = n.ToOutput()
	}

	return &model.NotificationListOutput{
		Data:        outputs,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
		UnreadCount: unread,
	}, nil
}

// MarkRead marks a single notification as read; enforces ownership.
func (s *NotificationService) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return s.notifRepo.MarkRead(ctx, notificationID, userID)
}

// MarkAllRead marks all of a user's notifications as read.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

// RegisterDeviceToken registers (or refreshes) a push-notification device token.
func (s *NotificationService) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token string, platform model.Platform) (*model.DeviceToken, error) {
	return s.notifRepo.RegisterDeviceToken(ctx, userID, token, platform)
}

// RevokeDeviceToken revokes a push-notification device token.
func (s *NotificationService) RevokeDeviceToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.notifRepo.RevokeDeviceToken(ctx, userID, token)
}
