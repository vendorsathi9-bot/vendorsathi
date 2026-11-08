package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/products/model"
	"vendorsaathi/internal/shared"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) CreateCategory(ctx context.Context, category *model.Category) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO categories (id, name, parent_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, category.ID, category.Name, category.ParentID, category.Active, category.CreatedAt, category.UpdatedAt)
	return err
}

func (r *ProductRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, parent_id, active, created_at, updated_at
		FROM categories WHERE id = $1
	`, id).Scan(&category.ID, &category.Name, &category.ParentID, &category.Active, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *ProductRepository) GetCategories(ctx context.Context, activeOnly bool) ([]model.Category, error) {
	query := `SELECT id, name, parent_id, active, created_at, updated_at FROM categories`
	args := []interface{}{}
	if activeOnly {
		query += ` WHERE active = TRUE`
	}
	query += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.Active, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *ProductRepository) UpdateCategory(ctx context.Context, id uuid.UUID, input model.UpdateCategoryInput) (*model.Category, error) {
	query := `
		UPDATE categories
		SET name = COALESCE(NULLIF($1, ''), name),
		    active = COALESCE($2, active),
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, parent_id, active, created_at, updated_at
	`
	var category model.Category
	err := r.pool.QueryRow(ctx, query, input.Name, input.Active, id).Scan(
		&category.ID, &category.Name, &category.ParentID, &category.Active, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO products (id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, product.ID, product.VendorID, product.CategoryID, product.Name, product.Description, product.PricePaise, product.CostPricePaise, product.Currency, product.ImageObjectKey, product.Active, product.CreatedAt, product.UpdatedAt)
	return err
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.pool.QueryRow(ctx, `
		SELECT id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at
		FROM products WHERE id = $1
	`, id).Scan(&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description, &product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

// ResolveVendorID maps an owner user ID or vendor profile ID to the vendor_profiles.id
func (r *ProductRepository) ResolveVendorID(ctx context.Context, idOrOwnerID uuid.UUID) uuid.UUID {
	var vendorID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM vendor_profiles WHERE id = $1 OR owner_user_id = $1`, idOrOwnerID).Scan(&vendorID)
	if err == nil {
		return vendorID
	}
	return idOrOwnerID
}

func (r *ProductRepository) GetProductsByVendorID(ctx context.Context, vendorID uuid.UUID, limit int, cursor string, activeOnly bool) ([]model.Product, string, bool, error) {
	query := `
		SELECT id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at
		FROM products WHERE vendor_id = $1
	`
	args := []interface{}{vendorID}
	argIdx := 2

	if activeOnly {
		query += ` AND active = TRUE`
	}

	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(argIdx+'0'))
	args = append(args, limit+1)
	argIdx++

	if cursor != "" {
		cursorID, err := uuid.Parse(cursor)
		if err != nil {
			return nil, "", false, err
		}
		query += ` AND id < $` + string(rune(argIdx+'0'))
		args = append(args, cursorID)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", false, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		if err := rows.Scan(&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description, &product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, "", false, err
		}
		products = append(products, product)
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	var nextCursor string
	if hasMore && len(products) > 0 {
		nextCursor = products[len(products)-1].ID.String()
	}

	return products, nextCursor, hasMore, rows.Err()
}

func (r *ProductRepository) GetActiveProductsByVendorID(ctx context.Context, vendorID uuid.UUID) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at
		FROM products WHERE vendor_id = $1 AND active = TRUE ORDER BY name
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		if err := rows.Scan(&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description, &product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, id uuid.UUID, input model.UpdateProductInput) (*model.Product, error) {
	query := `
		UPDATE products
		SET category_id = COALESCE($1, category_id),
		    name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE($3, description),
		    price_paise = COALESCE($4, price_paise),
		    cost_price_paise = COALESCE($5, cost_price_paise),
		    currency = COALESCE(NULLIF($6, ''), currency),
		    image_object_key = COALESCE($7, image_object_key),
		    active = COALESCE($8, active),
		    updated_at = NOW()
		WHERE id = $9
		RETURNING id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at
	`
	var product model.Product
	err := r.pool.QueryRow(ctx, query,
		input.CategoryID, input.Name, input.Description, input.PricePaise, input.CostPricePaise,
		input.Currency, input.ImageObjectKey, input.Active, id).Scan(
		&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description,
		&product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active,
		&product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) ArchiveProduct(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.pool.QueryRow(ctx, `
		UPDATE products SET active = FALSE, updated_at = NOW() WHERE id = $1
		RETURNING id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at
	`, id).Scan(&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description, &product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetProductsByCategoryID(ctx context.Context, categoryID uuid.UUID, activeOnly bool) ([]model.Product, error) {
	query := `SELECT id, vendor_id, category_id, name, description, price_paise, cost_price_paise, currency, image_object_key, active, created_at, updated_at FROM products WHERE category_id = $1`
	args := []interface{}{categoryID}
	if activeOnly {
		query += ` AND active = TRUE`
	}
	query += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		if err := rows.Scan(&product.ID, &product.VendorID, &product.CategoryID, &product.Name, &product.Description, &product.PricePaise, &product.CostPricePaise, &product.Currency, &product.ImageObjectKey, &product.Active, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}
