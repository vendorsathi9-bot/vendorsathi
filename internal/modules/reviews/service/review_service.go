package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/reviews/model"
	"vendorsaathi/internal/modules/reviews/repository"
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(ctx context.Context, customerID uuid.UUID, input model.CreateReviewInput) (*model.Review, error) {
	return s.repo.CreateReview(ctx, customerID, input)
}

func (s *ReviewService) GetVendorReviews(ctx context.Context, vendorID uuid.UUID) (*model.VendorReviewSummary, error) {
	return s.repo.GetVendorReviews(ctx, vendorID)
}

func (s *ReviewService) GetAllReviews(ctx context.Context) ([]model.Review, error) {
	return s.repo.GetAllReviews(ctx)
}

func (s *ReviewService) ModerateReview(ctx context.Context, reviewID uuid.UUID, isHidden bool) error {
	return s.repo.ModerateReview(ctx, reviewID, isHidden)
}

func (s *ReviewService) GetLoyaltySummary(ctx context.Context, customerID uuid.UUID) (*model.LoyaltySummaryOutput, error) {
	return s.repo.GetLoyaltySummary(ctx, customerID)
}
