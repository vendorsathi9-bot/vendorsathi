package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/carts/model"
	"vendorsaathi/internal/shared"
)

type CartRepository struct {
	pool *pgxpool.Pool
}

func NewCartRepository(pool *pgxpool.Pool) *CartRepository {
	return &CartRepository{pool: pool}
}

// GetOrCreateActiveCart returns the existing ACTIVE cart for a customer+vendor pair,
// or creates a new one if none exists.
func (r *CartRepository) GetOrCreateActiveCart(ctx context.Context, customerID, vendorID uuid.UUID) (*model.Cart, error) {
	var cart model.Cart
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, status, updated_at
		FROM carts
		WHERE customer_id = $1 AND vendor_id = $2 AND status = 'ACTIVE'
		LIMIT 1
	`, customerID, vendorID).Scan(&cart.ID, &cart.CustomerID, &cart.VendorID, &cart.Status, &cart.UpdatedAt)
	if err == nil {
		return &cart, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Create a new active cart
	newCart := &model.Cart{
		ID:         uuid.New(),
		CustomerID: customerID,
		VendorID:   vendorID,
		Status:     model.CartStatusActive,
		UpdatedAt:  time.Now().UTC(),
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO carts (id, customer_id, vendor_id, status, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, newCart.ID, newCart.CustomerID, newCart.VendorID, newCart.Status, newCart.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return newCart, nil
}

// GetCartByID returns a cart by ID, verifying customer ownership.
func (r *CartRepository) GetCartByID(ctx context.Context, cartID, customerID uuid.UUID) (*model.Cart, error) {
	var cart model.Cart
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, status, updated_at
		FROM carts WHERE id = $1 AND customer_id = $2
	`, cartID, customerID).Scan(&cart.ID, &cart.CustomerID, &cart.VendorID, &cart.Status, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &cart, nil
}

// GetActiveCart returns the customer's current active cart (any vendor).
func (r *CartRepository) GetActiveCart(ctx context.Context, customerID uuid.UUID) (*model.Cart, error) {
	var cart model.Cart
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, status, updated_at
		FROM carts WHERE customer_id = $1 AND status = 'ACTIVE'
		ORDER BY updated_at DESC
		LIMIT 1
	`, customerID).Scan(&cart.ID, &cart.CustomerID, &cart.VendorID, &cart.Status, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &cart, nil
}

// AddOrUpdateCartItem upserts a cart item (by cart_id + product_id).
func (r *CartRepository) AddOrUpdateCartItem(ctx context.Context, cartID, productID uuid.UUID, quantity int64) (*model.CartItem, error) {
	now := time.Now().UTC()
	var item model.CartItem
	err := r.pool.QueryRow(ctx, `
		INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = EXCLUDED.updated_at
		RETURNING id, cart_id, product_id, quantity, created_at, updated_at
	`, uuid.New(), cartID, productID, quantity, now).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// Touch the cart's updated_at
	_, _ = r.pool.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID)
	return &item, nil
}

// RemoveCartItem deletes a cart item by cart + product.
func (r *CartRepository) RemoveCartItem(ctx context.Context, cartID, productID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetCartItems returns all items for a cart, joined with product info.
func (r *CartRepository) GetCartItems(ctx context.Context, cartID uuid.UUID) ([]model.CartItemOutput, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ci.id, ci.cart_id, ci.product_id,
		       p.name AS product_name,
		       p.price_paise AS unit_price_paise,
		       ci.quantity,
		       (p.price_paise * ci.quantity) AS line_total_paise,
		       ci.updated_at
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at ASC
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItemOutput
	for rows.Next() {
		var item model.CartItemOutput
		if err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID,
			&item.ProductName, &item.UnitPricePaise, &item.Quantity,
			&item.LineTotalPaise, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []model.CartItemOutput{}
	}
	return items, rows.Err()
}

// ClearCart deletes all items from a cart.
func (r *CartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}

// UpdateCartStatus sets a new status on the cart.
func (r *CartRepository) UpdateCartStatus(ctx context.Context, cartID uuid.UUID, status model.CartStatus) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE carts SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, cartID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetCartWithItems returns a full cart with its items.
func (r *CartRepository) GetCartWithItems(ctx context.Context, cartID uuid.UUID) (*model.CartOutput, error) {
	var cart model.Cart
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, status, updated_at
		FROM carts WHERE id = $1
	`, cartID).Scan(&cart.ID, &cart.CustomerID, &cart.VendorID, &cart.Status, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	items, err := r.GetCartItems(ctx, cartID)
	if err != nil {
		return nil, err
	}

	return &model.CartOutput{
		ID:         cart.ID,
		CustomerID: cart.CustomerID,
		VendorID:   cart.VendorID,
		Status:     cart.Status,
		Items:      items,
		UpdatedAt:  cart.UpdatedAt,
	}, nil
}

// GetActiveCartForVendor returns the ACTIVE cart for a customer+vendor pair if it exists.
func (r *CartRepository) GetActiveCartForVendor(ctx context.Context, customerID, vendorID uuid.UUID) (*model.Cart, error) {
	var cart model.Cart
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, vendor_id, status, updated_at
		FROM carts WHERE customer_id = $1 AND vendor_id = $2 AND status = 'ACTIVE'
		LIMIT 1
	`, customerID, vendorID).Scan(&cart.ID, &cart.CustomerID, &cart.VendorID, &cart.Status, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &cart, nil
}
