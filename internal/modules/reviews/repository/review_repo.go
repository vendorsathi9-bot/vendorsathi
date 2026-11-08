package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/reviews/model"
	"vendorsaathi/internal/shared"
)

type ReviewRepository struct {
	pool *pgxpool.Pool
}

func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, customerID uuid.UUID, input model.CreateReviewInput) (*model.Review, error) {
	var review model.Review
	query := `
		INSERT INTO reviews (customer_id, vendor_id, order_id, rating, comment, is_hidden, created_at)
		VALUES ($1, $2, $3, $4, $5, FALSE, NOW())
		RETURNING id, customer_id, vendor_id, order_id, rating, comment, is_hidden, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		customerID, input.VendorID, input.OrderID, input.Rating, input.Comment,
	).Scan(
		&review.ID, &review.CustomerID, &review.VendorID, &review.OrderID,
		&review.Rating, &review.Comment, &review.IsHidden, &review.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetVendorReviews(ctx context.Context, vendorID uuid.UUID) (*model.VendorReviewSummary, error) {
	var summary model.VendorReviewSummary
	summary.Reviews = []model.Review{}

	// Aggregates
	aggQuery := `
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM reviews
		WHERE vendor_id = $1 AND is_hidden = FALSE
	`
	err := r.pool.QueryRow(ctx, aggQuery, vendorID).Scan(&summary.AverageRating, &summary.TotalReviews)
	if err != nil {
		return nil, err
	}

	// List
	listQuery := `
		SELECT r.id, r.customer_id, r.vendor_id, r.order_id, r.rating, r.comment, r.is_hidden, r.created_at,
		       COALESCE(cp.display_name, 'Customer') as customer_name
		FROM reviews r
		LEFT JOIN customer_profiles cp ON r.customer_id = cp.user_id
		WHERE r.vendor_id = $1 AND r.is_hidden = FALSE
		ORDER BY r.created_at DESC
		LIMIT 50
	`
	rows, err := r.pool.Query(ctx, listQuery, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rev model.Review
		if err := rows.Scan(
			&rev.ID, &rev.CustomerID, &rev.VendorID, &rev.OrderID,
			&rev.Rating, &rev.Comment, &rev.IsHidden, &rev.CreatedAt, &rev.CustomerName,
		); err != nil {
			return nil, err
		}
		summary.Reviews = append(summary.Reviews, rev)
	}

	return &summary, rows.Err()
}

func (r *ReviewRepository) GetAllReviews(ctx context.Context) ([]model.Review, error) {
	query := `
		SELECT r.id, r.customer_id, r.vendor_id, r.order_id, r.rating, r.comment, r.is_hidden, r.created_at,
		       COALESCE(cp.display_name, 'Customer') as customer_name
		FROM reviews r
		LEFT JOIN customer_profiles cp ON r.customer_id = cp.user_id
		ORDER BY r.created_at DESC
		LIMIT 100
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var rev model.Review
		if err := rows.Scan(
			&rev.ID, &rev.CustomerID, &rev.VendorID, &rev.OrderID,
			&rev.Rating, &rev.Comment, &rev.IsHidden, &rev.CreatedAt, &rev.CustomerName,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, rows.Err()
}

func (r *ReviewRepository) ModerateReview(ctx context.Context, reviewID uuid.UUID, isHidden bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE reviews SET is_hidden = $1 WHERE id = $2`, isHidden, reviewID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *ReviewRepository) GetLoyaltySummary(ctx context.Context, customerID uuid.UUID) (*model.LoyaltySummaryOutput, error) {
	var summary model.LoyaltySummaryOutput
	summary.Transactions = []model.LoyaltyTransaction{}

	var accountID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, points_balance FROM loyalty_accounts WHERE customer_id = $1
	`, customerID).Scan(&accountID, &summary.PointsBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			summary.PointsBalance = 0
			return &summary, nil
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, account_id, order_id, points, transaction_type, description, created_at
		FROM loyalty_transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tx model.LoyaltyTransaction
		if err := rows.Scan(&tx.ID, &tx.AccountID, &tx.OrderID, &tx.Points, &tx.TransactionType, &tx.Description, &tx.CreatedAt); err != nil {
			return nil, err
		}
		summary.Transactions = append(summary.Transactions, tx)
	}

	return &summary, rows.Err()
}
