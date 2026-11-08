package model

import (
	"time"

	"github.com/google/uuid"
)

type VendorSummary struct {
	TotalOrders               int64 `json:"total_orders"`
	CompletedOrders           int64 `json:"completed_orders"`
	CancelledOrders           int64 `json:"cancelled_orders"`
	PendingOrders             int64 `json:"pending_orders"`
	GrossSalesPaise           int64 `json:"gross_sales_paise"`
	ConfirmedCollectionsPaise int64 `json:"confirmed_collections_paise"`
	PendingCollectionsPaise   int64 `json:"pending_collections_paise"`
	OutstandingUdhaarPaise    int64 `json:"outstanding_udhaar_paise"`
}

type PlatformOverview struct {
	TotalUsers            int64 `json:"total_users"`
	TotalVendors          int64 `json:"total_vendors"`
	ActiveVendors         int64 `json:"active_vendors"`
	TotalCustomers        int64 `json:"total_customers"`
	TotalOrders           int64 `json:"total_orders"`
	CompletedOrders       int64 `json:"completed_orders"`
	TotalGrossSalesPaise  int64 `json:"total_gross_sales_paise"`
	TotalCollectionsPaise int64 `json:"total_collections_paise"`
	TotalProducts         int64 `json:"total_products"`
	TotalCategories       int64 `json:"total_categories"`
}

type ProductProfitMargin struct {
	ProductID        uuid.UUID `json:"product_id"`
	Name             string    `json:"name"`
	SellingPricePaise int64    `json:"selling_price_paise"`
	CostPricePaise   int64     `json:"cost_price_paise"`
	MarginPercent    float64   `json:"margin_percent"`
	UnitsSold        int64     `json:"units_sold"`
	RevenuePaise     int64     `json:"revenue_paise"`
	ProfitPaise      int64     `json:"profit_paise"`
}

type SalesTrendItem struct {
	Date        string `json:"date"`
	OrderCount  int64  `json:"order_count"`
	SalesPaise  int64  `json:"sales_paise"`
	ProfitPaise int64  `json:"profit_paise"`
}

type LowStockItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	Name         string    `json:"name"`
	OnHandQty    int64     `json:"on_hand_qty"`
	ReservedQty  int64     `json:"reserved_qty"`
	AvailableQty int64     `json:"available_qty"`
}

type BestSellingProduct struct {
	ProductID    uuid.UUID `json:"product_id"`
	Name         string    `json:"name"`
	UnitsSold    int64     `json:"units_sold"`
	RevenuePaise int64     `json:"revenue_paise"`
}

type Expense struct {
	ID          uuid.UUID `db:"id" json:"id"`
	VendorID    uuid.UUID `db:"vendor_id" json:"vendor_id"`
	Title       string    `db:"title" json:"title"`
	Category    string    `db:"category" json:"category"`
	AmountPaise int64     `db:"amount_paise" json:"amount_paise"`
	ExpenseDate string    `db:"expense_date" json:"expense_date"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type CreateExpenseInput struct {
	Title       string `json:"title" binding:"required"`
	Category    string `json:"category"`
	AmountPaise int64  `json:"amount_paise" binding:"required,min=1"`
	ExpenseDate string `json:"expense_date"`
}

type DetailedVendorAnalytics struct {
	DailySalesPaise           int64                 `json:"daily_sales_paise"`
	WeeklySalesPaise          int64                 `json:"weekly_sales_paise"`
	MonthlySalesPaise         int64                 `json:"monthly_sales_paise"`
	EstimatedGrossProfitPaise int64                 `json:"estimated_gross_profit_paise"`
	TotalExpensesPaise        int64                 `json:"total_expenses_paise"`
	EstimatedNetProfitPaise   int64                 `json:"estimated_net_profit_paise"`
	ProductProfitMargins      []ProductProfitMargin `json:"product_profit_margins"`
	BestSellingProducts       []BestSellingProduct  `json:"best_selling_products"`
	LowStockProducts          []LowStockItem        `json:"low_stock_products"`
	SalesTrends               []SalesTrendItem      `json:"sales_trends"`
}

type DemandForecastItem struct {
	ProductID          uuid.UUID `json:"product_id"`
	ProductName        string    `json:"product_name"`
	CurrentStock       int64     `json:"current_stock"`
	HistoricalDailyAvg float64   `json:"historical_daily_avg"`
	Predicted7DayDemand int64    `json:"predicted_7day_demand"`
	SuggestedRestockQty int64    `json:"suggested_restock_qty"`
	Confidence         string    `json:"confidence"` // HIGH, MEDIUM, LOW
	Explanation        string    `json:"explanation"`
}

type DemandForecastResponse struct {
	VendorID             uuid.UUID            `json:"vendor_id"`
	TotalProductsTracked int                  `json:"total_products_tracked"`
	HasSufficientHistory bool                 `json:"has_sufficient_history"`
	GeneralInsight       string               `json:"general_insight"`
	Forecasts            []DemandForecastItem `json:"forecasts"`
}

type WeatherResponse struct {
	City                 string   `json:"city"`
	TemperatureCelsius   float64  `json:"temperature_celsius"`
	Condition            string   `json:"condition"`
	ConditionDescription string   `json:"condition_description"`
	HumidityPercent      int      `json:"humidity_percent"`
	RainProbability      int      `json:"rain_probability"`
	WindSpeedKmh         float64  `json:"wind_speed_kmh"`
	Alerts               []string `json:"alerts"`
	VendorPlanningTips   []string `json:"vendor_planning_tips"`
}
