package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	cartsRepo "vendorsaathi/internal/modules/carts/repository"
	invModel "vendorsaathi/internal/modules/inventory/model"
	invRepo "vendorsaathi/internal/modules/inventory/repository"
	"vendorsaathi/internal/modules/orders/model"
	ordersRepo "vendorsaathi/internal/modules/orders/repository"
	productsRepo "vendorsaathi/internal/modules/products/repository"
	"vendorsaathi/internal/platform/logger"
	"vendorsaathi/internal/shared"
)

// validTransitions defines allowed state machine transitions.
// Key is the actor who performs the action; not enforced at this layer (handler enforces roles).
var validTransitions = map[model.OrderStatus][]model.OrderStatus{
	model.OrderStatusPending:               {model.OrderStatusAccepted, model.OrderStatusRejected, model.OrderStatusCancellationRequested},
	model.OrderStatusAccepted:              {model.OrderStatusPreparing, model.OrderStatusCancellationRequested},
	model.OrderStatusPreparing:             {model.OrderStatusReadyForPickup, model.OrderStatusOutForDelivery},
	model.OrderStatusReadyForPickup:        {model.OrderStatusCompleted},
	model.OrderStatusOutForDelivery:        {model.OrderStatusCompleted},
	model.OrderStatusCancellationRequested: {model.OrderStatusCancelled, model.OrderStatusPending, model.OrderStatusAccepted},
}

func isValidTransition(from, to model.OrderStatus) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

type OrderService struct {
	pool        *pgxpool.Pool
	orderRepo   *ordersRepo.OrderRepository
	cartRepo    *cartsRepo.CartRepository
	invRepo     *invRepo.InventoryRepository
	productRepo *productsRepo.ProductRepository
}

func NewOrderService(
	pool *pgxpool.Pool,
	orderRepo *ordersRepo.OrderRepository,
	cartRepo *cartsRepo.CartRepository,
	invRepo *invRepo.InventoryRepository,
	productRepo *productsRepo.ProductRepository,
) *OrderService {
	return &OrderService{
		pool:        pool,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		invRepo:     invRepo,
		productRepo: productRepo,
	}
}

// Checkout converts the customer's active cart into an order atomically.
func (s *OrderService) Checkout(ctx context.Context, customerID uuid.UUID, input model.CheckoutInput) (*model.OrderOutput, error) {
	// Validate delivery address requirement
	if input.FulfilmentType == model.FulfilmentTypeDelivery && input.DeliveryAddressSnapshot == nil {
		return nil, shared.NewValidationError("delivery_address_snapshot", "required when fulfilment_type is DELIVERY")
	}

	// Check idempotency before beginning the transaction
	if input.IdempotencyKey != "" {
		existingID, err := s.orderRepo.CheckIdempotency(ctx, customerID, input.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existingID != nil {
			return s.orderRepo.GetOrderByID(ctx, *existingID)
		}
	}

	// Load the active cart
	cart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.NewAppError("CART_EMPTY", "no active cart found", shared.ErrConflict)
		}
		return nil, err
	}

	// Get cart items (with product info)
	cartItems, err := s.cartRepo.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, shared.NewAppError("CART_EMPTY", "cart has no items", shared.ErrConflict)
	}

	vendorID := cart.VendorID

	// Collect sorted product IDs for deterministic lock ordering
	productIDs := make([]uuid.UUID, 0, len(cartItems))
	qtyMap := make(map[uuid.UUID]int64, len(cartItems))
	for _, item := range cartItems {
		productIDs = append(productIDs, item.ProductID)
		qtyMap[item.ProductID] = item.Quantity
	}
	sort.Slice(productIDs, func(i, j int) bool {
		return productIDs[i].String() < productIDs[j].String()
	})

	// Begin transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock inventory rows in sorted order to prevent deadlocks
	type inventoryRow struct {
		productID       uuid.UUID
		onHand          int64
		reserved        int64
		version         int64
		name            string
		pricePaise      int64
		active          bool
		productVendorID uuid.UUID
	}
	invRows := make([]inventoryRow, 0, len(productIDs))

	for _, pid := range productIDs {
		var row inventoryRow
		row.productID = pid

		// Lock the inventory row
		err := tx.QueryRow(ctx, `
			SELECT i.on_hand_qty, i.reserved_qty, i.version,
			       p.name, p.price_paise, p.active, p.vendor_id
			FROM inventory i
			JOIN products p ON i.product_id = p.id
			WHERE i.product_id = $1
			FOR UPDATE OF i
		`, pid).Scan(&row.onHand, &row.reserved, &row.version,
			&row.name, &row.pricePaise, &row.active, &row.productVendorID)
		if err != nil {
			return nil, fmt.Errorf("lock inventory for product %s: %w", pid, err)
		}

		if !row.active {
			return nil, shared.NewAppError("PRODUCT_INACTIVE", fmt.Sprintf("product %q is no longer available", row.name), shared.ErrConflict)
		}
		if row.productVendorID != vendorID {
			return nil, shared.NewForbidden("checkout items from this vendor")
		}

		available := row.onHand - row.reserved
		requested := qtyMap[pid]
		if requested > available {
			return nil, shared.NewInsufficientStock(available, requested)
		}

		invRows = append(invRows, row)
	}

	// Compute totals (server-side prices only)
	var subtotal int64
	orderItems := make([]model.OrderItem, 0, len(invRows))
	now := time.Now().UTC()
	orderID := uuid.New()

	for _, row := range invRows {
		qty := qtyMap[row.productID]
		lineTotal := row.pricePaise * qty
		subtotal += lineTotal

		pid := row.productID
		orderItems = append(orderItems, model.OrderItem{
			ID:                  uuid.New(),
			OrderID:             orderID,
			ProductID:           &pid,
			ProductNameSnapshot: row.name,
			UnitPricePaise:      row.pricePaise,
			Quantity:            qty,
			LineTotalPaise:      lineTotal,
			CreatedAt:           now,
		})
	}
	var discountPaise int64
	if input.RedeemLoyaltyPoints > 0 {
		var pointsBalance int
		var accountID uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id, points_balance FROM loyalty_accounts WHERE customer_id = $1 FOR UPDATE`, customerID).Scan(&accountID, &pointsBalance)
		if err == nil && pointsBalance >= input.RedeemLoyaltyPoints {
			requestedDiscount := int64(input.RedeemLoyaltyPoints) * 100
			if requestedDiscount > subtotal {
				discountPaise = subtotal
			} else {
				discountPaise = requestedDiscount
			}
			_, _ = tx.Exec(ctx, `
				UPDATE loyalty_accounts
				SET points_balance = points_balance - $1, updated_at = NOW()
				WHERE id = $2
			`, input.RedeemLoyaltyPoints, accountID)

			_, _ = tx.Exec(ctx, `
				INSERT INTO loyalty_transactions (account_id, order_id, points, transaction_type, description, created_at)
				VALUES ($1, $2, $3, 'REDEEMED', 'Redeemed during checkout', NOW())
				ON CONFLICT (account_id, order_id, transaction_type) DO NOTHING
			`, accountID, orderID, -input.RedeemLoyaltyPoints)
		}
	}

	total := subtotal - discountPaise
	if total < 0 {
		total = 0
	}

	var idemKey *string
	if input.IdempotencyKey != "" {
		k := input.IdempotencyKey
		idemKey = &k
	}

	order := &model.Order{
		ID:                      orderID,
		CustomerID:              customerID,
		VendorID:                vendorID,
		FulfilmentType:          input.FulfilmentType,
		Status:                  model.OrderStatusPending,
		SubtotalPaise:           subtotal,
		DiscountPaise:           discountPaise,
		TotalPaise:              total,
		DeliveryAddressSnapshot: input.DeliveryAddressSnapshot,
		IdempotencyKey:          idemKey,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Insert order + items + status history
	if err := s.orderRepo.CreateOrder(ctx, tx, order, orderItems); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Reserve inventory for each item (in sorted order, still inside the same tx)
	refType := "ORDER"
	for _, row := range invRows {
		qty := qtyMap[row.productID]
		newReserved := row.reserved + qty
		newVersion := row.version + 1

		_, err = tx.Exec(ctx, `
			UPDATE inventory
			SET reserved_qty = $1, version = $2, updated_at = NOW()
			WHERE product_id = $3 AND version = $4
		`, newReserved, newVersion, row.productID, row.version)
		if err != nil {
			return nil, fmt.Errorf("reserve inventory for product %s: %w", row.productID, err)
		}

		pid := row.productID
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements
			    (id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, uuid.New(), row.productID, invModel.MovementTypeReservation, -qty,
			nil, customerID, refType, orderID, now)
		if err != nil {
			return nil, fmt.Errorf("insert reservation movement for product %s: %w", pid, err)
		}
	}

	// Mark cart as CONVERTED (use the tx)
	_, err = tx.Exec(ctx, `UPDATE carts SET status = 'CONVERTED', updated_at = NOW() WHERE id = $1`, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("convert cart: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit checkout: %w", err)
	}

	// Post-commit: log for notification hook (actual notification module comes later)
	logger.Logger.Info().
		Str("order_id", orderID.String()).
		Str("customer_id", customerID.String()).
		Str("vendor_id", vendorID.String()).
		Msg("order created - notification trigger placeholder")

	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// GetOrderByID returns a full order. Callers must enforce ownership.
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.OrderOutput, error) {
	return s.orderRepo.GetOrderByID(ctx, id)
}

// GetOrdersByCustomer returns paginated orders for a customer.
func (s *OrderService) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	return s.orderRepo.GetOrdersByCustomer(ctx, customerID, limit, cursor)
}

// GetOrdersByVendor returns paginated orders for a vendor.
func (s *OrderService) GetOrdersByVendor(ctx context.Context, vendorID uuid.UUID, status string, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	return s.orderRepo.GetOrdersByVendor(ctx, vendorID, status, limit, cursor)
}

// GetAllOrders returns paginated orders for admin oversight.
func (s *OrderService) GetAllOrders(ctx context.Context, status string, limit int, cursor string) ([]model.OrderOutput, string, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.orderRepo.GetAllOrders(ctx, status, limit, cursor)
}

// transitionOrder is a generic helper that locks and transitions an order status.
func (s *OrderService) transitionOrder(ctx context.Context, orderID, actorID uuid.UUID, toStatus model.OrderStatus, reason *string, extraWork func(tx interface {
	Exec(ctx context.Context, sql string, args ...any) (interface{}, error)
}, order *model.Order) error) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}

	if !isValidTransition(order.Status, toStatus) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(toStatus))
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, toStatus, actorID, reason); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// AcceptOrder transitions PENDING → ACCEPTED.
func (s *OrderService) AcceptOrder(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	return s.transitionOrder(ctx, orderID, vendorUserID, model.OrderStatusAccepted, reason, nil)
}

// RejectOrder transitions PENDING → REJECTED and releases inventory.
func (s *OrderService) RejectOrder(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if !isValidTransition(order.Status, model.OrderStatusRejected) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(model.OrderStatusRejected))
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, model.OrderStatusRejected, vendorUserID, reason); err != nil {
		return nil, err
	}

	// Release inventory
	if err := s.releaseInventoryInTx(ctx, tx, orderID, vendorUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// PrepareOrder transitions ACCEPTED → PREPARING.
func (s *OrderService) PrepareOrder(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	return s.transitionOrder(ctx, orderID, vendorUserID, model.OrderStatusPreparing, reason, nil)
}

// MarkReadyForPickup transitions PREPARING → READY_FOR_PICKUP.
func (s *OrderService) MarkReadyForPickup(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if order.FulfilmentType != model.FulfilmentTypePickup {
		return nil, shared.NewAppError("INVALID_FULFILMENT", "ready-for-pickup only valid for PICKUP orders", shared.ErrInvalidTransition)
	}
	if !isValidTransition(order.Status, model.OrderStatusReadyForPickup) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(model.OrderStatusReadyForPickup))
	}
	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, model.OrderStatusReadyForPickup, vendorUserID, reason); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// MarkOutForDelivery transitions PREPARING → OUT_FOR_DELIVERY.
func (s *OrderService) MarkOutForDelivery(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if order.FulfilmentType != model.FulfilmentTypeDelivery {
		return nil, shared.NewAppError("INVALID_FULFILMENT", "out-for-delivery only valid for DELIVERY orders", shared.ErrInvalidTransition)
	}
	if !isValidTransition(order.Status, model.OrderStatusOutForDelivery) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(model.OrderStatusOutForDelivery))
	}
	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, model.OrderStatusOutForDelivery, vendorUserID, reason); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// CompleteOrder transitions READY_FOR_PICKUP or OUT_FOR_DELIVERY → COMPLETED and fulfils inventory.
func (s *OrderService) CompleteOrder(ctx context.Context, orderID, vendorUserID uuid.UUID, reason *string) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if !isValidTransition(order.Status, model.OrderStatusCompleted) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(model.OrderStatusCompleted))
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, model.OrderStatusCompleted, vendorUserID, reason); err != nil {
		return nil, err
	}

	// Fulfil inventory: decrement on_hand and reserved
	if err := s.fulfilInventoryInTx(ctx, tx, orderID, vendorUserID); err != nil {
		return nil, err
	}

	// Award loyalty points for completed order (1 point per ₹10, i.e., total_paise / 1000)
	earnedPoints := int(order.TotalPaise / 1000)
	if earnedPoints > 0 {
		var accountID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO loyalty_accounts (customer_id, points_balance, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (customer_id) DO UPDATE
			SET points_balance = loyalty_accounts.points_balance + $2, updated_at = NOW()
			RETURNING id
		`, order.CustomerID, earnedPoints).Scan(&accountID)
		if err == nil {
			_, _ = tx.Exec(ctx, `
				INSERT INTO loyalty_transactions (account_id, order_id, points, transaction_type, description, created_at)
				VALUES ($1, $2, $3, 'EARNED', 'Reward for completed purchase', NOW())
				ON CONFLICT (account_id, order_id, transaction_type) DO NOTHING
			`, accountID, orderID, earnedPoints)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// RequestCancellation transitions PENDING or ACCEPTED → CANCELLATION_REQUESTED.
func (s *OrderService) RequestCancellation(ctx context.Context, orderID, customerID uuid.UUID, reason string) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if order.CustomerID != customerID {
		return nil, shared.ErrForbidden
	}
	if !isValidTransition(order.Status, model.OrderStatusCancellationRequested) {
		return nil, shared.NewInvalidTransition(string(order.Status), string(model.OrderStatusCancellationRequested))
	}

	if err := s.orderRepo.CreateCancellationRequest(ctx, tx, orderID, customerID, reason, order.Status); err != nil {
		return nil, err
	}
	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, model.OrderStatusCancellationRequested, customerID, &reason); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// DecideCancellation lets the vendor approve or decline a cancellation request.
func (s *OrderService) DecideCancellation(ctx context.Context, orderID, vendorUserID uuid.UUID, input model.CancellationDecisionInput) (*model.OrderOutput, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.GetOrderForUpdate(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != model.OrderStatusCancellationRequested {
		return nil, shared.NewInvalidTransition(string(order.Status), "decide-cancellation")
	}

	cancellation, err := s.orderRepo.GetCancellationRequest(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdateCancellationDecision(ctx, tx, orderID, input.Decision, vendorUserID); err != nil {
		return nil, err
	}

	var nextStatus model.OrderStatus
	if input.Decision == model.CancellationDecisionApproved {
		nextStatus = model.OrderStatusCancelled
	} else {
		// Restore previous status
		nextStatus = model.OrderStatus(cancellation.PreviousStatus)
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, order.Status, nextStatus, vendorUserID, input.Reason); err != nil {
		return nil, err
	}

	// If approved: release inventory reservation
	if input.Decision == model.CancellationDecisionApproved {
		if err := s.releaseInventoryInTx(ctx, tx, orderID, vendorUserID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// releaseInventoryInTx releases reserved inventory for all items in an order, within a tx.
func (s *OrderService) releaseInventoryInTx(ctx context.Context, tx pgx.Tx, orderID, actorID uuid.UUID) error {
	return s.mutateInventoryInTx(ctx, tx, orderID, actorID, false)
}

// fulfilInventoryInTx fulfils inventory for all items in an order, within a tx.
func (s *OrderService) fulfilInventoryInTx(ctx context.Context, tx pgx.Tx, orderID, actorID uuid.UUID) error {
	return s.mutateInventoryInTx(ctx, tx, orderID, actorID, true)
}

// mutateInventoryInTx handles both release (fulfil=false) and fulfilment (fulfil=true).
func (s *OrderService) mutateInventoryInTx(ctx context.Context, tx pgx.Tx, orderID, actorID uuid.UUID, fulfil bool) error {
	rows, err := tx.Query(ctx, `
		SELECT product_id, quantity FROM order_items WHERE order_id = $1 AND product_id IS NOT NULL
	`, orderID)
	if err != nil {
		return err
	}

	type itemRow struct {
		productID uuid.UUID
		qty       int64
	}
	var items []itemRow
	for rows.Next() {
		var r itemRow
		if err := rows.Scan(&r.productID, &r.qty); err != nil {
			rows.Close()
			return err
		}
		items = append(items, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// Sort for deterministic lock ordering
	sort.Slice(items, func(i, j int) bool {
		return items[i].productID.String() < items[j].productID.String()
	})

	refType := "ORDER"
	for _, item := range items {
		var onHand, reserved, version int64
		err := tx.QueryRow(ctx, `
			SELECT on_hand_qty, reserved_qty, version FROM inventory WHERE product_id = $1 FOR UPDATE
		`, item.productID).Scan(&onHand, &reserved, &version)
		if err != nil {
			return fmt.Errorf("lock inventory product %s: %w", item.productID, err)
		}

		var newOnHand, newReserved int64
		var movType invModel.MovementType
		var qtyDelta int64

		if fulfil {
			newOnHand = onHand - item.qty
			newReserved = reserved - item.qty
			movType = invModel.MovementTypeFulfilment
			qtyDelta = -item.qty
		} else {
			newOnHand = onHand
			newReserved = reserved - item.qty
			movType = invModel.MovementTypeRelease
			qtyDelta = item.qty
		}

		_, err = tx.Exec(ctx, `
			UPDATE inventory
			SET on_hand_qty = $1, reserved_qty = $2, version = $3, updated_at = NOW()
			WHERE product_id = $4 AND version = $5
		`, newOnHand, newReserved, version+1, item.productID, version)
		if err != nil {
			return fmt.Errorf("update inventory product %s: %w", item.productID, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements
			    (id, product_id, movement_type, qty_delta, reason, actor_user_id, reference_type, reference_id, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		`, uuid.New(), item.productID, movType, qtyDelta, nil, actorID, refType, orderID)
		if err != nil {
			return fmt.Errorf("insert movement for product %s: %w", item.productID, err)
		}
	}
	return nil
}
