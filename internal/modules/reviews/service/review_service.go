package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/reviews/model"
	"vendorsaathi/internal/modules/reviews/repository"
	"vendorsaathi/internal/shared"
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(ctx context.Context, customerID uuid.UUID, input model.CreateReviewInput) (*model.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, shared.NewValidationError("rating", "must be between 1 and 5")
	}

	if input.OrderID != nil {
		vendorID, status, err := s.repo.OrderReviewEligibility(ctx, *input.OrderID, customerID)
		if err != nil {
			return nil, err
		}
		if vendorID != input.VendorID {
			return nil, shared.NewForbidden("review this vendor for the supplied order")
		}
		if status != "COMPLETED" {
			return nil, shared.NewAppError("REVIEW_NOT_ALLOWED", "only completed orders can be reviewed", shared.ErrConflict)
		}
		exists, err := s.repo.ReviewExistsForOrder(ctx, *input.OrderID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, shared.NewConflict("a review already exists for this order")
		}
	} else {
		// Order-less reviews are still restricted to vendors the customer has
		// actually purchased from.
		ok, err := s.repo.CustomerHasCompletedOrderWithVendor(ctx, customerID, input.VendorID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, shared.NewForbidden("review this vendor without a completed order")
		}
	}

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
