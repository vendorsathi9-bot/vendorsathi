package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/products/model"
	"vendorsaathi/internal/modules/products/repository"
	"vendorsaathi/internal/shared"
)

type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateCategory(ctx context.Context, input model.CreateCategoryInput) (*model.CategoryOutput, error) {
	if input.ParentID != nil {
		_, err := s.productRepo.GetCategoryByID(ctx, *input.ParentID)
		if err != nil {
			return nil, err
		}
	}

	category := &model.Category{
		ID:        uuid.New(),
		Name:      input.Name,
		ParentID:  input.ParentID,
		Active:    true,
		CreatedAt: shared.Now(),
		UpdatedAt: shared.Now(),
	}

	if err := s.productRepo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}

	output := category.ToOutput()
	return &output, nil
}

func (s *ProductService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.CategoryOutput, error) {
	category, err := s.productRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	output := category.ToOutput()
	return &output, nil
}

func (s *ProductService) GetCategories(ctx context.Context, activeOnly bool) ([]model.CategoryOutput, error) {
	categories, err := s.productRepo.GetCategories(ctx, activeOnly)
	if err != nil {
		return nil, err
	}
	outputs := make([]model.CategoryOutput, len(categories))
	for i, c := range categories {
		outputs[i] = c.ToOutput()
	}
	return outputs, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, id uuid.UUID, input model.UpdateCategoryInput) (*model.CategoryOutput, error) {
	updated, err := s.productRepo.UpdateCategory(ctx, id, input)
	if err != nil {
		return nil, err
	}
	output := updated.ToOutput()
	return &output, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, vendorID uuid.UUID, input model.CreateProductInput) (*model.ProductOutput, error) {
	vendorID = s.productRepo.ResolveVendorID(ctx, vendorID)
	if input.CategoryID != nil {
		_, err := s.productRepo.GetCategoryByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}
	}

	currency := "INR"
	if input.Currency != nil {
		currency = *input.Currency
	}

	var costPrice int64
	if input.CostPricePaise != nil {
		costPrice = *input.CostPricePaise
	}

	product := &model.Product{
		ID:             uuid.New(),
		VendorID:       vendorID,
		CategoryID:     input.CategoryID,
		Name:           input.Name,
		Description:    input.Description,
		PricePaise:     input.PricePaise,
		CostPricePaise: costPrice,
		Currency:       currency,
		ImageObjectKey: input.ImageObjectKey,
		Active:         true,
		CreatedAt:      shared.Now(),
		UpdatedAt:      shared.Now(),
	}

	if err := s.productRepo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	output := product.ToOutput()
	return &output, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*model.ProductOutput, error) {
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	output := product.ToOutput()
	return &output, nil
}

func (s *ProductService) GetMyProducts(ctx context.Context, vendorID uuid.UUID, limit int, cursor string, activeOnly bool) (*model.ProductListOutput, error) {
	vendorID = s.productRepo.ResolveVendorID(ctx, vendorID)
	products, nextCursor, hasMore, err := s.productRepo.GetProductsByVendorID(ctx, vendorID, limit, cursor, activeOnly)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.ProductOutput, len(products))
	for i, p := range products {
		outputs[i] = p.ToOutput()
	}

	return &model.ProductListOutput{
		Data:       outputs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *ProductService) GetMyActiveProducts(ctx context.Context, vendorID uuid.UUID) ([]model.ProductOutput, error) {
	vendorID = s.productRepo.ResolveVendorID(ctx, vendorID)
	products, err := s.productRepo.GetActiveProductsByVendorID(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.ProductOutput, len(products))
	for i, p := range products {
		outputs[i] = p.ToOutput()
	}
	return outputs, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, vendorID uuid.UUID, productID uuid.UUID, input model.UpdateProductInput) (*model.ProductOutput, error) {
	vendorID = s.productRepo.ResolveVendorID(ctx, vendorID)
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if product.VendorID != vendorID {
		return nil, shared.ErrNotFound
	}

	if input.CategoryID != nil {
		_, err := s.productRepo.GetCategoryByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}
	}

	updated, err := s.productRepo.UpdateProduct(ctx, productID, input)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *ProductService) ArchiveProduct(ctx context.Context, vendorID uuid.UUID, productID uuid.UUID) (*model.ProductOutput, error) {
	vendorID = s.productRepo.ResolveVendorID(ctx, vendorID)
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if product.VendorID != vendorID {
		return nil, shared.ErrNotFound
	}

	archived, err := s.productRepo.ArchiveProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	output := archived.ToOutput()
	return &output, nil
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, activeOnly bool) ([]model.ProductOutput, error) {
	products, err := s.productRepo.GetProductsByCategoryID(ctx, categoryID, activeOnly)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.ProductOutput, len(products))
	for i, p := range products {
		outputs[i] = p.ToOutput()
	}
	return outputs, nil
}

func (s *ProductService) GetVendorProducts(ctx context.Context, vendorID uuid.UUID) ([]model.ProductOutput, error) {
	resolvedID := s.productRepo.ResolveVendorID(ctx, vendorID)
	products, err := s.productRepo.GetActiveProductsByVendorID(ctx, resolvedID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.ProductOutput, len(products))
	for i, p := range products {
		outputs[i] = p.ToOutput()
	}
	return outputs, nil
}
