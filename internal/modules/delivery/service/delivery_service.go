package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/delivery/model"
	"vendorsaathi/internal/modules/delivery/repository"
	"vendorsaathi/internal/shared"
)

type DeliveryService struct {
	repo *repository.DeliveryRepository
}

func NewDeliveryService(repo *repository.DeliveryRepository) *DeliveryService {
	return &DeliveryService{repo: repo}
}

// AssignDelivery assigns an agent to an order. Called by Admin.
func (s *DeliveryService) AssignDelivery(ctx context.Context, orderID, agentUserID, adminUserID uuid.UUID) (*model.DeliveryAssignmentOutput, error) {
	a, err := s.repo.AssignDelivery(ctx, orderID, agentUserID, adminUserID)
	if err != nil {
		return nil, err
	}
	out := a.ToOutput()
	return &out, nil
}

// GetAssignmentByOrderID returns assignment for a specific order.
// Access: relevant customer, vendor owner, the agent, or admin.
func (s *DeliveryService) GetAssignmentByOrderID(ctx context.Context, orderID, actorID uuid.UUID, actorRoles []string) (*model.DeliveryAssignmentOutput, error) {
	a, err := s.repo.GetAssignmentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if !s.canAccessOrderDelivery(ctx, orderID, actorID, actorRoles, a) {
		return nil, shared.NewForbidden("access this delivery information")
	}

	out := a.ToOutput()
	return &out, nil
}

// GetMyAssignments returns all assignments for the calling delivery agent.
func (s *DeliveryService) GetMyAssignments(ctx context.Context, agentUserID uuid.UUID) ([]model.DeliveryAssignmentWithOrderOutput, error) {
	assignments, err := s.repo.GetAssignmentsByAgent(ctx, agentUserID)
	if err != nil {
		return nil, err
	}

	out := make([]model.DeliveryAssignmentWithOrderOutput, len(assignments))
	for i, a := range assignments {
		out[i] = a.ToOutput()
	}
	return out, nil
}

// UpdateDeliveryStatus updates the delivery status. Only the assigned agent can do this.
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, orderID, agentUserID uuid.UUID, newStatus model.DeliveryStatus) (*model.DeliveryAssignmentOutput, error) {
	current, err := s.repo.GetAssignmentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if current.AgentUserID != agentUserID {
		return nil, shared.NewForbidden("update delivery status: not assigned agent")
	}

	if !current.Status.IsValidTransition(newStatus) {
		return nil, shared.NewInvalidTransition(string(current.Status), string(newStatus))
	}

	a, err := s.repo.UpdateAssignmentStatus(ctx, orderID, newStatus)
	if err != nil {
		return nil, err
	}
	out := a.ToOutput()
	return &out, nil
}

// UpdateLocation upserts the agent's GPS coordinates. Only the assigned agent can call this.
func (s *DeliveryService) UpdateLocation(ctx context.Context, orderID, agentUserID uuid.UUID, lat, lng float64) (*model.DeliveryTrackingOutput, error) {
	assignment, err := s.repo.GetAssignmentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if assignment.AgentUserID != agentUserID {
		return nil, shared.NewForbidden("update location: not assigned agent")
	}
	if assignment.Status == model.DeliveryStatusDelivered {
		return nil, shared.NewValidationError("delivery_status", "delivery is already completed")
	}

	t, err := s.repo.UpsertLocation(ctx, orderID, agentUserID, lat, lng)
	if err != nil {
		return nil, err
	}
	out := t.ToOutput()
	return &out, nil
}

// GetLocation returns the latest agent location. Access: relevant customer, vendor, agent, admin.
func (s *DeliveryService) GetLocation(ctx context.Context, orderID, actorID uuid.UUID, actorRoles []string) (*model.DeliveryTrackingOutput, error) {
	assignment, err := s.repo.GetAssignmentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	if !s.canAccessOrderDelivery(ctx, orderID, actorID, actorRoles, assignment) {
		return nil, shared.NewForbidden("access tracking for this order")
	}

	t, err := s.repo.GetLocation(ctx, orderID)
	if err != nil {
		return nil, err
	}
	out := t.ToOutput()
	return &out, nil
}

// ListAllAssignments returns all assignments for admin view.
func (s *DeliveryService) ListAllAssignments(ctx context.Context) ([]model.DeliveryAssignmentWithOrderOutput, error) {
	assignments, err := s.repo.ListAllAssignments(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]model.DeliveryAssignmentWithOrderOutput, len(assignments))
	for i, a := range assignments {
		out[i] = a.ToOutput()
	}
	return out, nil
}

// canAccessOrderDelivery checks if the actor is authorized to view delivery info.
func (s *DeliveryService) canAccessOrderDelivery(ctx context.Context, orderID, actorID uuid.UUID, actorRoles []string, assignment *model.DeliveryAssignment) bool {
	for _, r := range actorRoles {
		if r == "ADMIN" || r == "SUPER_ADMIN" {
			return true
		}
	}

	if assignment.AgentUserID == actorID {
		return true
	}

	customerID, vendorOwnerID, err := s.repo.GetOrderMeta(ctx, orderID)
	if err != nil {
		return false
	}
	return actorID == customerID || actorID == vendorOwnerID
}
