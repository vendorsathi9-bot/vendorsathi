package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/audit/model"
	"vendorsaathi/internal/modules/audit/repository"
	"vendorsaathi/internal/platform/logger"
)

// AuditService owns audit-log business logic.
type AuditService struct {
	auditRepo *repository.AuditRepository
}

// NewAuditService constructs an AuditService.
func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// Log creates an audit log entry. Errors are swallowed and logged to stderr so
// that callers are never failed by audit instrumentation.
func (s *AuditService) Log(ctx context.Context, input model.CreateAuditLogInput) {
	if _, err := s.auditRepo.Create(ctx, input); err != nil {
		logger.Logger.Error().Err(err).
			Str("action", input.Action).
			Str("resource_type", input.ResourceType).
			Msg("audit: failed to write audit log")
	}
}

// List returns a paginated list of audit logs filtered by optional criteria.
// This should only be called by admin-authorised handlers.
func (s *AuditService) List(ctx context.Context, actorUserID *uuid.UUID, resourceType, resourceID *string, limit int, cursor string) (*model.AuditLogListOutput, error) {
	filters := repository.AuditListFilters{
		ActorUserID: actorUserID,
	}
	if resourceType != nil && *resourceType != "" {
		filters.ResourceType = resourceType
	}
	if resourceID != nil && *resourceID != "" {
		rid, err := uuid.Parse(*resourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid resource_id: %w", err)
		}
		filters.ResourceID = &rid
	}

	if limit <= 0 {
		limit = 20
	}

	logs, nextCursor, hasMore, err := s.auditRepo.List(ctx, filters, limit, cursor)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.AuditLogOutput, len(logs))
	for i, l := range logs {
		outputs[i] = l.ToOutput()
	}

	return &model.AuditLogListOutput{
		Data:       outputs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
