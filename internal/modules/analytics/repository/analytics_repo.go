package repository

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/analytics/model"
	"vendorsaathi/internal/shared"
)

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) GetVendorProfileByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (uuid.UUID, error) {
	var vendorID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM vendor_profiles WHERE id = $1 OR owner_user_id = $1`, ownerUserID).Scan(&vendorID)
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

func (r *AnalyticsRepository) GetDetailedVendorAnalytics(ctx context.Context, vendorID uuid.UUID) (*model.DetailedVendorAnalytics, error) {
	var a model.DetailedVendorAnalytics
	a.ProductProfitMargins = []model.ProductProfitMargin{}
	a.BestSellingProducts = []model.BestSellingProduct{}
	a.LowStockProducts = []model.LowStockItem{}
	a.SalesTrends = []model.SalesTrendItem{}

	// 1. Daily, Weekly, Monthly sales
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_paise), 0) FROM orders
		WHERE vendor_id = $1 AND status = 'COMPLETED' AND created_at >= CURRENT_DATE
	`, vendorID).Scan(&a.DailySalesPaise)

	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_paise), 0) FROM orders
		WHERE vendor_id = $1 AND status = 'COMPLETED' AND created_at >= NOW() - INTERVAL '7 days'
	`, vendorID).Scan(&a.WeeklySalesPaise)

	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_paise), 0) FROM orders
		WHERE vendor_id = $1 AND status = 'COMPLETED' AND created_at >= NOW() - INTERVAL '30 days'
	`, vendorID).Scan(&a.MonthlySalesPaise)

	// 2. Gross profit from completed orders
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM((oi.unit_price_paise - COALESCE(p.cost_price_paise, 0)) * oi.quantity), 0)
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		LEFT JOIN products p ON oi.product_id = p.id
		WHERE o.vendor_id = $1 AND o.status = 'COMPLETED'
	`, vendorID).Scan(&a.EstimatedGrossProfitPaise)

	// 3. Total expenses
	_ = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_paise), 0) FROM expenses WHERE vendor_id = $1
	`, vendorID).Scan(&a.TotalExpensesPaise)

	// 4. Net profit
	a.EstimatedNetProfitPaise = a.EstimatedGrossProfitPaise - a.TotalExpensesPaise

	// 5. Product profit margins
	marginQuery := `
		SELECT p.id, p.name, p.price_paise, COALESCE(p.cost_price_paise, 0),
		       COALESCE(SUM(oi.quantity), 0) as units_sold,
		       COALESCE(SUM(oi.line_total_paise), 0) as revenue_paise,
		       COALESCE(SUM((oi.unit_price_paise - COALESCE(p.cost_price_paise, 0)) * oi.quantity), 0) as profit_paise
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id AND o.status = 'COMPLETED'
		WHERE p.vendor_id = $1
		GROUP BY p.id, p.name, p.price_paise, p.cost_price_paise
		ORDER BY revenue_paise DESC
		LIMIT 20
	`
	marginRows, err := r.pool.Query(ctx, marginQuery, vendorID)
	if err == nil {
		defer marginRows.Close()
		for marginRows.Next() {
			var m model.ProductProfitMargin
			if err := marginRows.Scan(&m.ProductID, &m.Name, &m.SellingPricePaise, &m.CostPricePaise, &m.UnitsSold, &m.RevenuePaise, &m.ProfitPaise); err == nil {
				if m.SellingPricePaise > 0 {
					m.MarginPercent = float64(m.SellingPricePaise-m.CostPricePaise) / float64(m.SellingPricePaise) * 100
					if m.MarginPercent < 0 {
						m.MarginPercent = 0
					}
				}
				a.ProductProfitMargins = append(a.ProductProfitMargins, m)
				if len(a.BestSellingProducts) < 5 && m.UnitsSold > 0 {
					a.BestSellingProducts = append(a.BestSellingProducts, model.BestSellingProduct{
						ProductID:    m.ProductID,
						Name:         m.Name,
						UnitsSold:    m.UnitsSold,
						RevenuePaise: m.RevenuePaise,
					})
				}
			}
		}
	}

	// 6. Low stock products
	stockRows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, i.on_hand_qty, i.reserved_qty, (i.on_hand_qty - i.reserved_qty) as available_qty
		FROM products p
		JOIN inventory i ON p.id = i.product_id
		WHERE p.vendor_id = $1 AND (i.on_hand_qty - i.reserved_qty) <= 5
		ORDER BY available_qty ASC
		LIMIT 10
	`, vendorID)
	if err == nil {
		defer stockRows.Close()
		for stockRows.Next() {
			var ls model.LowStockItem
			if err := stockRows.Scan(&ls.ProductID, &ls.Name, &ls.OnHandQty, &ls.ReservedQty, &ls.AvailableQty); err == nil {
				a.LowStockProducts = append(a.LowStockProducts, ls)
			}
		}
	}

	// 7. Sales trends (last 7 days)
	trendQuery := `
		SELECT to_char(d.day, 'YYYY-MM-DD') as date_str,
		       COUNT(DISTINCT o.id) as order_count,
		       COALESCE(SUM(o.total_paise), 0) as sales_paise,
		       COALESCE(SUM((oi.unit_price_paise - COALESCE(p.cost_price_paise, 0)) * oi.quantity), 0) as profit_paise
		FROM generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day'::interval) d(day)
		LEFT JOIN orders o ON date_trunc('day', o.created_at) = d.day AND o.vendor_id = $1 AND o.status = 'COMPLETED'
		LEFT JOIN order_items oi ON o.id = oi.order_id
		LEFT JOIN products p ON oi.product_id = p.id
		GROUP BY d.day
		ORDER BY d.day ASC
	`
	trendRows, err := r.pool.Query(ctx, trendQuery, vendorID)
	if err == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var t model.SalesTrendItem
			if err := trendRows.Scan(&t.Date, &t.OrderCount, &t.SalesPaise, &t.ProfitPaise); err == nil {
				a.SalesTrends = append(a.SalesTrends, t)
			}
		}
	}

	return &a, nil
}

func (r *AnalyticsRepository) GetDemandForecast(ctx context.Context, vendorID uuid.UUID) (*model.DemandForecastResponse, error) {
	resp := &model.DemandForecastResponse{
		VendorID:  vendorID,
		Forecasts: []model.DemandForecastItem{},
	}

	// Check total distinct order days for this vendor
	var distinctDays int
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT date_trunc('day', created_at))
		FROM orders
		WHERE vendor_id = $1 AND status = 'COMPLETED'
	`, vendorID).Scan(&distinctDays)

	hasSufficient := distinctDays >= 3
	resp.HasSufficientHistory = hasSufficient

	if hasSufficient {
		resp.GeneralInsight = "AI Forecast is calibrated based on verified multi-day order cadence and current inventory stock levels."
	} else {
		resp.GeneralInsight = "Limited historical sales data detected (< 3 active days). Showing moving-average baseline demand. Precision will improve with more completed orders."
	}

	// Query each product and its total units sold over the last 14 days
	query := `
		SELECT p.id, p.name,
		       COALESCE(i.on_hand_qty - i.reserved_qty, 0) as available_qty,
		       COALESCE(SUM(oi.quantity), 0) as units_sold_recent,
		       GREATEST(1, COUNT(DISTINCT date_trunc('day', o.created_at))) as active_days
		FROM products p
		LEFT JOIN inventory i ON p.id = i.product_id
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id AND o.status = 'COMPLETED' AND o.created_at >= NOW() - INTERVAL '14 days'
		WHERE p.vendor_id = $1 AND p.active = TRUE
		GROUP BY p.id, p.name, i.on_hand_qty, i.reserved_qty
		ORDER BY p.name ASC
	`
	rows, err := r.pool.Query(ctx, query, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.DemandForecastItem
		var unitsSold int64
		var activeDays int

		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &unitsSold, &activeDays); err != nil {
			return nil, err
		}

		// Simple statistical forecasting:
		// Daily average velocity
		divisor := float64(activeDays)
		if divisor < 1 {
			divisor = 1
		}
		item.HistoricalDailyAvg = float64(unitsSold) / divisor
		// 7-day projected demand
		item.Predicted7DayDemand = int64(math.Ceil(item.HistoricalDailyAvg * 7))
		if item.Predicted7DayDemand == 0 && unitsSold > 0 {
			item.Predicted7DayDemand = 1
		}
		// Suggested restock quantity = max(0, predicted_demand - current_stock)
		diff := item.Predicted7DayDemand - item.CurrentStock
		if diff > 0 {
			item.SuggestedRestockQty = diff
		} else {
			item.SuggestedRestockQty = 0
		}

		if hasSufficient && unitsSold > 0 {
			item.Confidence = "HIGH"
			item.Explanation = "Consistent customer purchase frequency observed over active trading days."
		} else if unitsSold > 0 {
			item.Confidence = "MEDIUM"
			item.Explanation = "Initial sales detected. Run-rate projected based on early transactions."
		} else {
			item.Confidence = "LOW"
			item.Explanation = "No recent sales recorded. Stock maintaining current baseline levels."
		}

		resp.Forecasts = append(resp.Forecasts, item)
	}

	resp.TotalProductsTracked = len(resp.Forecasts)
	return resp, nil
}

func (r *AnalyticsRepository) CreateExpense(ctx context.Context, vendorID uuid.UUID, input model.CreateExpenseInput) (*model.Expense, error) {
	var exp model.Expense
	category := input.Category
	if category == "" {
		category = "OPERATIONAL"
	}
	query := `
		INSERT INTO expenses (vendor_id, title, category, amount_paise, expense_date, created_at)
		VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5, '')::date, CURRENT_DATE), NOW())
		RETURNING id, vendor_id, title, category, amount_paise, expense_date::text, created_at
	`
	err := r.pool.QueryRow(ctx, query, vendorID, input.Title, category, input.AmountPaise, input.ExpenseDate).Scan(
		&exp.ID, &exp.VendorID, &exp.Title, &exp.Category, &exp.AmountPaise, &exp.ExpenseDate, &exp.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *AnalyticsRepository) GetExpenses(ctx context.Context, vendorID uuid.UUID) ([]model.Expense, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, vendor_id, title, category, amount_paise, expense_date::text, created_at
		FROM expenses
		WHERE vendor_id = $1
		ORDER BY expense_date DESC, created_at DESC
		LIMIT 50
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []model.Expense{}
	for rows.Next() {
		var exp model.Expense
		if err := rows.Scan(&exp.ID, &exp.VendorID, &exp.Title, &exp.Category, &exp.AmountPaise, &exp.ExpenseDate, &exp.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}
	return expenses, rows.Err()
}

func (r *AnalyticsRepository) DeleteExpense(ctx context.Context, vendorID, expenseID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM expenses WHERE id = $1 AND vendor_id = $2`, expenseID, vendorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
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

	// Products and Categories
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&overview.TotalProducts)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories`).Scan(&overview.TotalCategories)

	return &overview, nil
}
