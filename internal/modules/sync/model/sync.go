package model

import (
	"time"

	"github.com/google/uuid"
)

// SyncOperationStatus represents the lifecycle state of a sync operation.
type SyncOperationStatus string

const (
	SyncStatusPending     SyncOperationStatus = "PENDING"
	SyncStatusSending     SyncOperationStatus = "SENDING"
	SyncStatusAccepted    SyncOperationStatus = "ACCEPTED"
	SyncStatusRejected    SyncOperationStatus = "REJECTED"
	SyncStatusNeedsReview SyncOperationStatus = "NEEDS_REVIEW"
)

// Allowlisted offline operation types.
const (
	OpTypeProductDescriptionUpdate = "PRODUCT_DESCRIPTION_UPDATE"
	OpTypeStockAdjustmentRequest   = "STOCK_ADJUSTMENT_REQUEST"
)

// AllowedOperationTypes is the set of operation types accepted by the sync endpoint.
var AllowedOperationTypes = map[string]bool{
	OpTypeProductDescriptionUpdate: true,
	OpTypeStockAdjustmentRequest:   true,
}

// SyncOperation is the domain object for an offline sync operation.
type SyncOperation struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	DeviceID          string
	ClientOperationID uuid.UUID
	OperationType     string
	PayloadVersion    int
	Status            SyncOperationStatus
	Payload           map[string]interface{}
	Result            map[string]interface{}
	CreatedAt         time.Time
	ProcessedAt       *time.Time
}

// SubmitSyncOperationInput is the parsed request body for submitting a sync op.
type SubmitSyncOperationInput struct {
	ClientOperationID uuid.UUID              `json:"client_operation_id" binding:"required"`
	DeviceID          string                 `json:"device_id" binding:"required"`
	OperationType     string                 `json:"operation_type" binding:"required"`
	PayloadVersion    int                    `json:"payload_version"`
	Payload           map[string]interface{} `json:"payload" binding:"required"`
}

// SyncOperationOutput is the JSON-serialisable form of a sync operation.
type SyncOperationOutput struct {
	ID                uuid.UUID              `json:"id"`
	UserID            uuid.UUID              `json:"user_id"`
	DeviceID          string                 `json:"device_id"`
	ClientOperationID uuid.UUID              `json:"client_operation_id"`
	OperationType     string                 `json:"operation_type"`
	PayloadVersion    int                    `json:"payload_version"`
	Status            SyncOperationStatus    `json:"status"`
	Payload           map[string]interface{} `json:"payload"`
	Result            map[string]interface{} `json:"result,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	ProcessedAt       *time.Time             `json:"processed_at,omitempty"`
}

// SyncOperationListOutput wraps a paginated list of sync-operation outputs.
type SyncOperationListOutput struct {
	Data       []SyncOperationOutput `json:"data"`
	NextCursor string                `json:"next_cursor,omitempty"`
	HasMore    bool                  `json:"has_more"`
}

// ToOutput converts a SyncOperation domain object to its serialisable form.
func (s *SyncOperation) ToOutput() SyncOperationOutput {
	return SyncOperationOutput{
		ID:                s.ID,
		UserID:            s.UserID,
		DeviceID:          s.DeviceID,
		ClientOperationID: s.ClientOperationID,
		OperationType:     s.OperationType,
		PayloadVersion:    s.PayloadVersion,
		Status:            s.Status,
		Payload:           s.Payload,
		Result:            s.Result,
		CreatedAt:         s.CreatedAt,
		ProcessedAt:       s.ProcessedAt,
	}
}
