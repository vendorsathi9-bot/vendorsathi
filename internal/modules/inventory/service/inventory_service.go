package service

import (
	"context"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/inventory/model"
	"vendorsaathi/internal/modules/inventory/repository"
)

type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
}

func NewInventoryService(inventoryRepo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

func (s *InventoryService) GetInventory(ctx context.Context, productID uuid.UUID) (*model.InventoryOutput, error) {
	inv, err := s.inventoryRepo.GetInventory(ctx, productID)
	if err != nil {
		return nil, err
	}
	output := inv.ToOutput()
	return &output, nil
}

func (s *InventoryService) GetInventoryByVendor(ctx context.Context, vendorID uuid.UUID) ([]model.InventoryOutput, error) {
	inventories, err := s.inventoryRepo.GetInventoryByVendor(ctx, vendorID)
	if err != nil {
		return nil, err
	}
	outputs := make([]model.InventoryOutput, len(inventories))
	for i, inv := range inventories {
		outputs[i] = inv.ToOutput()
	}
	return outputs, nil
}

func (s *InventoryService) AdjustStock(ctx context.Context, vendorID uuid.UUID, input model.AdjustStockInput) (*model.InventoryOutput, *model.InventoryMovementOutput, error) {
	inv, movement, err := s.inventoryRepo.AdjustStock(ctx, input.ProductID, input.QtyDelta, input.Reason, &vendorID, input.IdempotencyKey)
	if err != nil {
		return nil, nil, err
	}
	invOutput := inv.ToOutput()
	movOutput := movement.ToOutput()
	return &invOutput, &movOutput, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, vendorID uuid.UUID, input model.ReserveStockInput) (*model.InventoryOutput, *model.InventoryMovementOutput, error) {
	inv, movement, err := s.inventoryRepo.ReserveStock(ctx, input.ProductID, input.Qty, input.IdempotencyKey, input.ReferenceType, input.ReferenceID, &vendorID)
	if err != nil {
		return nil, nil, err
	}
	invOutput := inv.ToOutput()
	movOutput := movement.ToOutput()
	return &invOutput, &movOutput, nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, vendorID uuid.UUID, input model.ReleaseStockInput) (*model.InventoryOutput, *model.InventoryMovementOutput, error) {
	inv, movement, err := s.inventoryRepo.ReleaseStock(ctx, input.ProductID, input.Qty, input.IdempotencyKey, input.ReferenceType, input.ReferenceID, &vendorID)
	if err != nil {
		return nil, nil, err
	}
	invOutput := inv.ToOutput()
	movOutput := movement.ToOutput()
	return &invOutput, &movOutput, nil
}

func (s *InventoryService) FulfilStock(ctx context.Context, vendorID uuid.UUID, input model.FulfilStockInput) (*model.InventoryOutput, *model.InventoryMovementOutput, error) {
	inv, movement, err := s.inventoryRepo.FulfilStock(ctx, input.ProductID, input.Qty, input.IdempotencyKey, input.ReferenceType, input.ReferenceID, &vendorID)
	if err != nil {
		return nil, nil, err
	}
	invOutput := inv.ToOutput()
	movOutput := movement.ToOutput()
	return &invOutput, &movOutput, nil
}

func (s *InventoryService) GetMovements(ctx context.Context, productID uuid.UUID, limit int, cursor string) (*model.InventoryMovementListOutput, error) {
	movements, nextCursor, hasMore, err := s.inventoryRepo.GetMovements(ctx, productID, limit, cursor)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.InventoryMovementOutput, len(movements))
	for i, m := range movements {
		outputs[i] = m.ToOutput()
	}

	return &model.InventoryMovementListOutput{
		Data:       outputs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
