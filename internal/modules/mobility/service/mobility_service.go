package service

import (
	"context"

	"vendorsaathi/internal/modules/mobility/model"
	"vendorsaathi/internal/modules/mobility/repository"
)

type MobilityService struct {
	repo *repository.MobilityRepository
}

func NewMobilityService(repo *repository.MobilityRepository) *MobilityService {
	return &MobilityService{repo: repo}
}

func (s *MobilityService) FindNearbyVendors(ctx context.Context, lat, lng, radiusKm float64) ([]model.NearbyVendor, error) {
	if radiusKm <= 0 {
		radiusKm = 5.0
	}
	if radiusKm > 50.0 {
		radiusKm = 50.0
	}
	return s.repo.FindNearbyVendors(ctx, lat, lng, radiusKm)
}
