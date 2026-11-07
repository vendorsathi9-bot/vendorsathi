package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/users/model"
	"vendorsaathi/internal/modules/users/repository"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/shared"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) GetUser(ctx context.Context, actorID, targetID uuid.UUID, isAdmin bool) (*model.UserOutput, error) {
	if !isAdmin && actorID != targetID {
		return nil, shared.NewForbidden("view other user's profile")
	}

	user, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, targetID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, input model.UpdateProfileInput) (*model.UserOutput, error) {
	if input.Phone != "" {
		existing, err := s.userRepo.GetByPhone(ctx, input.Phone)
		if err == nil && existing.ID != userID {
			return nil, shared.NewConflict("phone already in use")
		}
		if !errors.Is(err, shared.ErrNotFound) && err != nil {
			return nil, err
		}
	}

	if input.Email != "" {
		existing, err := s.userRepo.GetByEmail(ctx, input.Email)
		if err == nil && existing.ID != userID {
			return nil, shared.NewConflict("email already in use")
		}
		if !errors.Is(err, shared.ErrNotFound) && err != nil {
			return nil, err
		}
	}

	user, err := s.userRepo.UpdateProfile(ctx, userID, input.Phone, input.Email)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) UpdateStatus(ctx context.Context, actorID, targetID uuid.UUID, isAdmin bool, input model.UpdateStatusInput) (*model.UserOutput, error) {
	if !isAdmin {
		return nil, shared.NewForbidden("update user status")
	}

	if actorID == targetID && input.Status == model.UserStatusSuspended {
		return nil, shared.NewConflict("cannot suspend yourself")
	}

	if err := s.userRepo.UpdateStatus(ctx, targetID, input.Status); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, targetID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	valid, err := auth.VerifyPassword(currentPassword, user.PasswordHash, 64*1024, 3, 4, 16, 32)
	if err != nil || !valid {
		return shared.NewAppError("INVALID_PASSWORD", "current password is incorrect", shared.ErrUnauthorized)
	}

	newHash, err := auth.HashPassword(newPassword, 64*1024, 3, 4, 16, 32)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, newHash)
}

func (s *UserService) AssignRole(ctx context.Context, actorID, targetID uuid.UUID, isAdmin bool, input model.AssignRoleInput) (*model.UserOutput, error) {
	if !isAdmin {
		return nil, shared.NewForbidden("assign roles")
	}

	role, err := s.userRepo.GetRoleByCode(ctx, input.RoleCode)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.AssignRole(ctx, targetID, role.ID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, targetID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) RevokeRole(ctx context.Context, actorID, targetID uuid.UUID, isAdmin bool, input model.RevokeRoleInput) (*model.UserOutput, error) {
	if !isAdmin {
		return nil, shared.NewForbidden("revoke roles")
	}

	if actorID == targetID {
		return nil, shared.NewConflict("cannot revoke your own roles")
	}

	role, err := s.userRepo.GetRoleByCode(ctx, input.RoleCode)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.RevokeRole(ctx, targetID, role.ID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(ctx, targetID)
	if err != nil {
		return nil, err
	}

	output := user.ToOutput(roles)
	return &output, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit int, cursor string, status *model.UserStatus) (*model.UserListOutput, error) {
	users, nextCursor, hasMore, err := s.userRepo.ListUsers(ctx, limit, cursor, status)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.UserOutput, len(users))
	for i, u := range users {
		roles, err := s.userRepo.GetUserRoles(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		outputs[i] = u.ToOutput(roles)
	}

	return &model.UserListOutput{
		Users:      outputs,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
