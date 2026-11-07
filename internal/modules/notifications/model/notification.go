package model

import (
	"time"

	"github.com/google/uuid"
)

// Platform represents the push-token device platform.
type Platform string

const (
	PlatformAndroid Platform = "ANDROID"
	PlatformIOS     Platform = "IOS"
	PlatformWeb     Platform = "WEB"
)

// Notification is the domain object for an in-app notification.
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	EventType string
	Title     string
	Body      string
	Payload   map[string]interface{}
	ReadAt    *time.Time
	CreatedAt time.Time
}

// DeviceToken represents a push-notification device registration.
type DeviceToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Token      string
	Platform   Platform
	LastSeenAt time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

// RegisterDeviceTokenInput is the request body for device-token registration.
type RegisterDeviceTokenInput struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=ANDROID IOS WEB"`
}

// NotificationOutput is the JSON-serialisable notification response.
type NotificationOutput struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	EventType string                 `json:"event_type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationListOutput wraps a paginated list of notifications with the
// caller's current unread count.
type NotificationListOutput struct {
	Data        []NotificationOutput `json:"data"`
	NextCursor  string               `json:"next_cursor,omitempty"`
	HasMore     bool                 `json:"has_more"`
	UnreadCount int64                `json:"unread_count"`
}

// ToOutput converts a Notification domain object to its serialisable form.
func (n *Notification) ToOutput() NotificationOutput {
	return NotificationOutput{
		ID:        n.ID,
		UserID:    n.UserID,
		EventType: n.EventType,
		Title:     n.Title,
		Body:      n.Body,
		Payload:   n.Payload,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
