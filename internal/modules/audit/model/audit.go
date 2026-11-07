package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a single audit trail entry.
type AuditLog struct {
	ID           uuid.UUID
	ActorUserID  *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Reason       *string
	RequestID    *uuid.UUID
	Metadata     map[string]interface{}
	CreatedAt    time.Time
}

// CreateAuditLogInput is passed to the service when logging an action.
type CreateAuditLogInput struct {
	ActorUserID  *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Reason       *string
	RequestID    *uuid.UUID
	Metadata     map[string]interface{}
}

// AuditLogOutput is the JSON-serialisable form returned by the API.
type AuditLogOutput struct {
	ID           uuid.UUID              `json:"id"`
	ActorUserID  *uuid.UUID             `json:"actor_user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Reason       *string                `json:"reason,omitempty"`
	RequestID    *uuid.UUID             `json:"request_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// AuditLogListOutput wraps a paginated list of audit-log outputs.
type AuditLogListOutput struct {
	Data       []AuditLogOutput `json:"data"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
}

// ToOutput converts an AuditLog domain object to its serialisable output form.
func (a *AuditLog) ToOutput() AuditLogOutput {
	return AuditLogOutput{
		ID:           a.ID,
		ActorUserID:  a.ActorUserID,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		Reason:       a.Reason,
		RequestID:    a.RequestID,
		Metadata:     a.Metadata,
		CreatedAt:    a.CreatedAt,
	}
}
