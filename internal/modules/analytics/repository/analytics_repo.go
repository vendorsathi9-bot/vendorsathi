package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/analytics/model"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) GetVendorProfileByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (uuid.UUID, error) {
	var vendorID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM vendor_profiles WHERE owner_user_id = $1`, ownerUserID).Scan(&vendorID)
	return vendorID, err
}

func (r *AnalyticsRepository) GetVendorSummary(ctx context.Context, vendorID uuid.UUID) (*model.VendorSummary, error) {
	var summary model.VendorSummary

	// Orders summary
	orderQuery := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'COMPLETED'),
			COUNT(*) FILTER (WHERE status IN ('CANCELLED', 'REJECTED')),
			COUNT(*) FILTER (WHERE status = 'PENDING'),
			COALESCE(SUM(total_paise) FILTER (WHERE status = 'COMPLETED'), 0)
		FROM orders
		WHERE vendor_id = $1;
	`
	err := r.pool.QueryRow(ctx, orderQuery, vendorID).Scan(
		&summary.TotalOrders,
		&summary.CompletedOrders,
		&summary.CancelledOrders,
		&summary.PendingOrders,
		&summary.GrossSalesPaise,
	)
	if err != nil {
		return nil, err
	}

	// Payments collections
	payQuery := `
		SELECT 
			COALESCE(SUM(p.amount_paise) FILTER (WHERE p.status = 'CONFIRMED'), 0),
			COALESCE(SUM(p.amount_paise) FILTER (WHERE p.status = 'PENDING_VERIFICATION'), 0)
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE o.vendor_id = $1;
	`
	err = r.pool.QueryRow(ctx, payQuery, vendorID).Scan(
		&summary.ConfirmedCollectionsPaise,
		&summary.PendingCollectionsPaise,
	)
	if err != nil {
		return nil, err
	}

	// Outstanding Udhaar for vendor
	udhaarQuery := `
		SELECT COALESCE(SUM(
			CASE ue.entry_type
				WHEN 'CREDIT' THEN ue.amount_paise
				WHEN 'ADJUSTMENT' THEN ue.amount_paise
				WHEN 'REPAYMENT' THEN -ue.amount_paise
				ELSE 0
			END
		), 0)
		FROM udhaar_entries ue
		JOIN udhaar_accounts ua ON ue.account_id = ua.id
		WHERE ua.vendor_id = $1 AND ua.status = 'ACTIVE';
	`
	err = r.pool.QueryRow(ctx, udhaarQuery, vendorID).Scan(&summary.OutstandingUdhaarPaise)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *AnalyticsRepository) GetPlatformOverview(ctx context.Context) (*model.PlatformOverview, error) {
	var overview model.PlatformOverview

	// Users counts
	userQuery := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE r.code = 'VENDOR'),
			COUNT(*) FILTER (WHERE r.code = 'CUSTOMER')
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id;
	`
	_ = r.pool.QueryRow(ctx, userQuery).Scan(
		&overview.TotalUsers,
		&overview.TotalVendors,
		&overview.TotalCustomers,
	)

	// Active vendors
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM vendor_profiles WHERE status = 'ACTIVE'`).Scan(&overview.ActiveVendors)

	// Orders and Sales
	orderQuery := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'COMPLETED'),
			COALESCE(SUM(total_paise) FILTER (WHERE status = 'COMPLETED'), 0)
		FROM orders;
	`
	_ = r.pool.QueryRow(ctx, orderQuery).Scan(
		&overview.TotalOrders,
		&overview.CompletedOrders,
		&overview.TotalGrossSalesPaise,
	)

	// Collections
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_paise), 0) FROM payments WHERE status = 'CONFIRMED'`).Scan(&overview.TotalCollectionsPaise)

	return &overview, nil
}
