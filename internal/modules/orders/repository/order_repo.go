package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/orders/model"
	"vendorsaathi/internal/shared"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// CreateOrder inserts an order, its items, and an initial PENDING status history row,
// all within the provided transaction.
func (r *OrderRepository) CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order, items []model.OrderItem) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO orders (id, customer_id, vendor_id, fulfilment_type, status,
		                    subtotal_paise, discount_paise, total_paise,
		                    delivery_address_snapshot, pickup_address_snapshot,
		                    idempotency_key, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, order.ID, order.CustomerID, order.VendorID, order.FulfilmentType, order.Status,
		order.SubtotalPaise, order.DiscountPaise, order.TotalPaise,
		order.DeliveryAddressSnapshot, order.PickupAddressSnapshot,
		order.IdempotencyKey, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, product_name_snapshot,
			                        unit_price_paise, quantity, line_total_paise, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, item.ID, item.OrderID, item.ProductID, item.ProductNameSnapshot,
			item.UnitPricePaise, item.Quantity, item.LineTotalPaise, item.CreatedAt)
		if err != nil {
			return err
		}
	}

	// Insert initial PENDING status history
	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, actor_user_id, reason, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, uuid.New(), order.ID, nil, string(model.OrderStatusPending), order.CustomerID, nil, order.CreatedAt)
	return err
}

// GetOrderByID returns a full order with its items.
func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.OrderOutput, error) {
	var o model.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, fulfilment_type, status,
		       subtotal_paise, discount_paise, total_paise,
		       delivery_address_snapshot, pickup_address_snapshot,
		       idempotency_key, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(&o.ID, &o.CustomerID, &o.VendorID, &o.FulfilmentType, &o.Status,
		&o.SubtotalPaise, &o.DiscountPaise, &o.TotalPaise,
		&o.DeliveryAddressSnapshot, &o.PickupAddressSnapshot,
		&o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	items, err := r.getOrderItems(ctx, r.pool, id)
	if err != nil {
		return nil, err
	}

	return toOrderOutput(&o, items), nil
}

// getOrderItems is a helper that queries order_items using a QueryExecutor interface.
func (r *OrderRepository) getOrderItems(ctx context.Context, q interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, orderID uuid.UUID) ([]model.OrderItemOutput, error) {
	rows, err := q.Query(ctx, `
		SELECT id, order_id, product_id, product_name_snapshot,
		       unit_price_paise, quantity, line_total_paise, created_at
		FROM order_items WHERE order_id = $1
		ORDER BY created_at ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItemOutput
	for rows.Next() {
		var item model.OrderItemOutput
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductNameSnapshot,
			&item.UnitPricePaise, &item.Quantity, &item.LineTotalPaise, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []model.OrderItemOutput{}
	}
	return items, rows.Err()
}

// GetOrdersByCustomer returns paginated orders for a customer.
func (r *OrderRepository) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	args := []interface{}{customerID, limit + 1}
	query := `
		SELECT id, customer_id, vendor_id, fulfilment_type, status,
		       subtotal_paise, discount_paise, total_paise,
		       delivery_address_snapshot, pickup_address_snapshot,
		       idempotency_key, created_at, updated_at
		FROM orders WHERE customer_id = $1
	`
	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, "", false, err
		}
		query += ` AND created_at < $3`
		args = append(args, cursorTime)
	}
	query += ` ORDER BY created_at DESC LIMIT $2`

	return r.queryOrders(ctx, query, args, limit)
}

// GetAllOrders returns paginated orders for administrative oversight.
func (r *OrderRepository) GetAllOrders(ctx context.Context, status string, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	args := []interface{}{limit + 1}
	query := `
		SELECT id, customer_id, vendor_id, fulfilment_type, status,
		       subtotal_paise, discount_paise, total_paise,
		       delivery_address_snapshot, pickup_address_snapshot,
		       idempotency_key, created_at, updated_at
		FROM orders WHERE 1=1
	`
	argIdx := 2
	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, "", false, err
		}
		query += ` AND created_at < $` + itoa(argIdx)
		args = append(args, cursorTime)
	}
	query += ` ORDER BY created_at DESC LIMIT $1`

	return r.queryOrders(ctx, query, args, limit)
}

// GetOrdersByVendor returns paginated orders for a vendor, optionally filtered by status.
func (r *OrderRepository) GetOrdersByVendor(ctx context.Context, vendorID uuid.UUID, status string, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	args := []interface{}{vendorID, limit + 1}
	query := `
		SELECT id, customer_id, vendor_id, fulfilment_type, status,
		       subtotal_paise, discount_paise, total_paise,
		       delivery_address_snapshot, pickup_address_snapshot,
		       idempotency_key, created_at, updated_at
		FROM orders WHERE (vendor_id = $1 OR vendor_id IN (SELECT id FROM vendor_profiles WHERE owner_user_id = $1))
	`
	argIdx := 3
	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, "", false, err
		}
		query += ` AND created_at < $` + itoa(argIdx)
		args = append(args, cursorTime)
	}
	query += ` ORDER BY created_at DESC LIMIT $2`

	return r.queryOrders(ctx, query, args, limit)
}

func (r *OrderRepository) queryOrders(ctx context.Context, query string, args []interface{}, limit int) ([]model.OrderOutput, string, bool, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.VendorID, &o.FulfilmentType, &o.Status,
			&o.SubtotalPaise, &o.DiscountPaise, &o.TotalPaise,
			&o.DeliveryAddressSnapshot, &o.PickupAddressSnapshot,
			&o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, "", false, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, "", false, err
	}

	hasMore := len(orders) > limit
	if hasMore {
		orders = orders[:limit]
	}

	var nextCursor string
	if hasMore && len(orders) > 0 {
		nextCursor = orders[len(orders)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	// Build outputs (without items for list view)
	outputs := make([]model.OrderOutput, 0, len(orders))
	for _, o := range orders {
		o := o
		outputs = append(outputs, *toOrderOutput(&o, []model.OrderItemOutput{}))
	}
	return outputs, nextCursor, hasMore, nil
}

// UpdateOrderStatus updates order status and appends a status history record, within a transaction.
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, fromStatus, toStatus model.OrderStatus, actorID uuid.UUID, reason *string) error {
	tag, err := tx.Exec(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
	`, toStatus, orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	fromStr := string(fromStatus)
	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, actor_user_id, reason, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,NOW())
	`, uuid.New(), orderID, fromStr, string(toStatus), actorID, reason)
	return err
}

// GetOrderForUpdate returns an order using SELECT FOR UPDATE within a transaction.
func (r *OrderRepository) GetOrderForUpdate(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) (*model.Order, error) {
	var o model.Order
	err := tx.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, fulfilment_type, status,
		       subtotal_paise, discount_paise, total_paise,
		       delivery_address_snapshot, pickup_address_snapshot,
		       idempotency_key, created_at, updated_at
		FROM orders WHERE id = $1 FOR UPDATE
	`, orderID).Scan(&o.ID, &o.CustomerID, &o.VendorID, &o.FulfilmentType, &o.Status,
		&o.SubtotalPaise, &o.DiscountPaise, &o.TotalPaise,
		&o.DeliveryAddressSnapshot, &o.PickupAddressSnapshot,
		&o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

// CreateCancellationRequest inserts a cancellation request within a transaction.
func (r *OrderRepository) CreateCancellationRequest(ctx context.Context, tx pgx.Tx, orderID, requestedBy uuid.UUID, reason string, previousStatus model.OrderStatus) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO order_cancellations (id, order_id, requested_by, reason, previous_status, created_at)
		VALUES ($1,$2,$3,$4,$5,NOW())
	`, uuid.New(), orderID, requestedBy, reason, string(previousStatus))
	return err
}

// GetCancellationRequest returns the cancellation request for an order.
func (r *OrderRepository) GetCancellationRequest(ctx context.Context, orderID uuid.UUID) (*model.OrderCancellation, error) {
	var c model.OrderCancellation
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, requested_by, reason, previous_status, decision, decided_by, decided_at, created_at
		FROM order_cancellations WHERE order_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, orderID).Scan(&c.ID, &c.OrderID, &c.RequestedBy, &c.Reason, &c.PreviousStatus,
		&c.Decision, &c.DecidedBy, &c.DecidedAt, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// UpdateCancellationDecision sets decision on a cancellation request, within a transaction.
func (r *OrderRepository) UpdateCancellationDecision(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, decision model.CancellationDecision, decidedBy uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE order_cancellations
		SET decision = $1, decided_by = $2, decided_at = $3
		WHERE order_id = $4 AND decision IS NULL
	`, decision, decidedBy, now, orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// CheckIdempotency returns the existing order ID if an order with the given idempotency key exists.
func (r *OrderRepository) CheckIdempotency(ctx context.Context, customerID uuid.UUID, idempotencyKey string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM orders WHERE customer_id = $1 AND idempotency_key = $2
	`, customerID, idempotencyKey).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

// GetOrderItemsByOrderID returns the raw OrderItem domain models for a given order (used in service logic).
func (r *OrderRepository) GetOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, product_id, product_name_snapshot,
		       unit_price_paise, quantity, line_total_paise, created_at
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductNameSnapshot,
			&item.UnitPricePaise, &item.Quantity, &item.LineTotalPaise, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// --- helpers ---

func toOrderOutput(o *model.Order, items []model.OrderItemOutput) *model.OrderOutput {
	return &model.OrderOutput{
		ID:                      o.ID,
		CustomerID:              o.CustomerID,
		VendorID:                o.VendorID,
		FulfilmentType:          o.FulfilmentType,
		Status:                  o.Status,
		SubtotalPaise:           o.SubtotalPaise,
		DiscountPaise:           o.DiscountPaise,
		TotalPaise:              o.TotalPaise,
		DeliveryAddressSnapshot: o.DeliveryAddressSnapshot,
		PickupAddressSnapshot:   o.PickupAddressSnapshot,
		Items:                   items,
		CreatedAt:               o.CreatedAt,
		UpdatedAt:               o.UpdatedAt,
	}
}

func itoa(i int) string {
	return string(rune('0' + i))
}
