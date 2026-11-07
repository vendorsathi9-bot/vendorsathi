package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	ParentID  *uuid.UUID `db:"parent_id"`
	Active    bool       `db:"active"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type Product struct {
	ID             uuid.UUID  `db:"id"`
	VendorID       uuid.UUID  `db:"vendor_id"`
	CategoryID     *uuid.UUID `db:"category_id"`
	Name           string     `db:"name"`
	Description    *string    `db:"description"`
	PricePaise     int64      `db:"price_paise"`
	Currency       string     `db:"currency"`
	ImageObjectKey *string    `db:"image_object_key"`
	Active         bool       `db:"active"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

type CreateCategoryInput struct {
	Name     string     `json:"name" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type UpdateCategoryInput struct {
	Name   *string `json:"name,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

type CreateProductInput struct {
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Name           string     `json:"name" binding:"required"`
	Description    *string    `json:"description,omitempty"`
	PricePaise     int64      `json:"price_paise" binding:"required,min=0"`
	Currency       *string    `json:"currency,omitempty"`
	ImageObjectKey *string    `json:"image_object_key,omitempty"`
}

type UpdateProductInput struct {
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Name           *string    `json:"name,omitempty"`
	Description    *string    `json:"description,omitempty"`
	PricePaise     *int64     `json:"price_paise,omitempty"`
	Currency       *string    `json:"currency,omitempty"`
	ImageObjectKey *string    `json:"image_object_key,omitempty"`
	Active         *bool      `json:"active,omitempty"`
}

type CategoryOutput struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ProductOutput struct {
	ID             uuid.UUID  `json:"id"`
	VendorID       uuid.UUID  `json:"vendor_id"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	PricePaise     int64      `json:"price_paise"`
	Currency       string     `json:"currency"`
	ImageObjectKey *string    `json:"image_object_key,omitempty"`
	Active         bool       `json:"active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ProductListOutput struct {
	Data       []ProductOutput `json:"data"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

func (c *Category) ToOutput() CategoryOutput {
	return CategoryOutput{
		ID:        c.ID,
		Name:      c.Name,
		ParentID:  c.ParentID,
		Active:    c.Active,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (p *Product) ToOutput() ProductOutput {
	return ProductOutput{
		ID:             p.ID,
		VendorID:       p.VendorID,
		CategoryID:     p.CategoryID,
		Name:           p.Name,
		Description:    p.Description,
		PricePaise:     p.PricePaise,
		Currency:       p.Currency,
		ImageObjectKey: p.ImageObjectKey,
		Active:         p.Active,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
