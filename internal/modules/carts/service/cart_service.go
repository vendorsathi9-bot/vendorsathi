package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/carts/model"
	cartsRepo "vendorsaathi/internal/modules/carts/repository"
	productsRepo "vendorsaathi/internal/modules/products/repository"
	"vendorsaathi/internal/shared"
)

type CartService struct {
	cartRepo    *cartsRepo.CartRepository
	productRepo *productsRepo.ProductRepository
}

func NewCartService(cartRepo *cartsRepo.CartRepository, productRepo *productsRepo.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// GetMyCart returns the customer's current active cart (any vendor).
func (s *CartService) GetMyCart(ctx context.Context, customerID uuid.UUID) (*model.CartOutput, error) {
	cart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			// No active cart: return empty representation
			return &model.CartOutput{
				CustomerID: customerID,
				Items:      []model.CartItemOutput{},
			}, nil
		}
		return nil, err
	}
	return s.cartRepo.GetCartWithItems(ctx, cart.ID)
}

// AddItem adds or updates an item in the customer's cart.
// Enforces the single-vendor rule: if the customer has an active cart for a different vendor, return conflict.
func (s *CartService) AddItem(ctx context.Context, customerID, productID uuid.UUID, quantity int64) (*model.CartOutput, error) {
	// Validate product exists and is active
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !product.Active {
		return nil, shared.NewAppError("PRODUCT_INACTIVE", "product is not available", shared.ErrConflict)
	}

	// Enforce single-vendor rule: check if there's an active cart for a different vendor
	existingCart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}
	if existingCart != nil && existingCart.VendorID != product.VendorID {
		return nil, shared.NewConflict("cannot add items from a different vendor; clear your existing cart first")
	}

	// Get or create an active cart for the product's vendor
	cart, err := s.cartRepo.GetOrCreateActiveCart(ctx, customerID, product.VendorID)
	if err != nil {
		return nil, err
	}

	_, err = s.cartRepo.AddOrUpdateCartItem(ctx, cart.ID, productID, quantity)
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetCartWithItems(ctx, cart.ID)
}

// RemoveItem removes an item from the customer's active cart.
func (s *CartService) RemoveItem(ctx context.Context, customerID, productID uuid.UUID) (*model.CartOutput, error) {
	cart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.RemoveCartItem(ctx, cart.ID, productID); err != nil {
		return nil, err
	}

	return s.cartRepo.GetCartWithItems(ctx, cart.ID)
}

// UpdateItemQuantity updates the quantity of an item in the customer's active cart.
func (s *CartService) UpdateItemQuantity(ctx context.Context, customerID, productID uuid.UUID, quantity int64) (*model.CartOutput, error) {
	cart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil {
		return nil, err
	}

	_, err = s.cartRepo.AddOrUpdateCartItem(ctx, cart.ID, productID, quantity)
	if err != nil {
		return nil, err
	}

	return s.cartRepo.GetCartWithItems(ctx, cart.ID)
}

// ClearCart clears all items from the customer's active cart.
func (s *CartService) ClearCart(ctx context.Context, customerID uuid.UUID) error {
	cart, err := s.cartRepo.GetActiveCart(ctx, customerID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil // nothing to clear
		}
		return err
	}
	return s.cartRepo.ClearCart(ctx, cart.ID)
}

// GetCart returns a specific cart with ownership check.
func (s *CartService) GetCart(ctx context.Context, customerID, cartID uuid.UUID) (*model.CartOutput, error) {
	_, err := s.cartRepo.GetCartByID(ctx, cartID, customerID)
	if err != nil {
		return nil, err
	}
	return s.cartRepo.GetCartWithItems(ctx, cartID)
}
