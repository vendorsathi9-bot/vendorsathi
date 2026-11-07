package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/analytics/model"
	"vendorsaathi/internal/modules/analytics/repository"
	"vendorsaathi/internal/shared"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsService(repo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetVendorSummary(ctx context.Context, vendorUserID uuid.UUID) (*model.VendorSummary, error) {
	vendorID, err := s.repo.GetVendorProfileByOwnerUserID(ctx, vendorUserID)
	if err != nil {
		return nil, shared.NewForbidden("user does not have an active vendor profile")
	}
	return s.repo.GetVendorSummary(ctx, vendorID)
}

func (s *AnalyticsService) GetPlatformOverview(ctx context.Context) (*model.PlatformOverview, error) {
	return s.repo.GetPlatformOverview(ctx)
}
