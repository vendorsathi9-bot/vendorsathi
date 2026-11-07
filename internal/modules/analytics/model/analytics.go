package model

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
}
