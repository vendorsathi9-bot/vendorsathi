package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/customers/model"
	"vendorsaathi/internal/modules/customers/repository"
	"vendorsaathi/internal/shared"
)

type CustomerService struct {
	customerRepo *repository.CustomerRepository
}

func NewCustomerService(customerRepo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (s *CustomerService) GetOrCreateProfile(ctx context.Context, userID uuid.UUID) (*model.CustomerProfileOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			profile = &model.CustomerProfile{
				UserID:             userID,
				ContactPreferences: []byte("{}"),
				CreatedAt:          shared.Now(),
				UpdatedAt:          shared.Now(),
			}
			if err := s.customerRepo.CreateProfile(ctx, profile); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	output := profile.ToOutput()
	return &output, nil
}

func (s *CustomerService) GetMyProfile(ctx context.Context, userID uuid.UUID) (*model.CustomerProfileOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	output := profile.ToOutput()
	return &output, nil
}

func (s *CustomerService) UpdateMyProfile(ctx context.Context, userID uuid.UUID, input model.UpdateCustomerProfileInput) (*model.CustomerProfileOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			profile = &model.CustomerProfile{
				UserID:             userID,
				ContactPreferences: []byte("{}"),
				CreatedAt:          shared.Now(),
				UpdatedAt:          shared.Now(),
			}
			if err := s.customerRepo.CreateProfile(ctx, profile); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	updated, err := s.customerRepo.UpdateProfile(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *CustomerService) CreateAddress(ctx context.Context, userID uuid.UUID, input model.CreateCustomerAddressInput) (*model.CustomerAddressOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			profile = &model.CustomerProfile{
				UserID:             userID,
				ContactPreferences: []byte("{}"),
				CreatedAt:          shared.Now(),
				UpdatedAt:          shared.Now(),
			}
			if err := s.customerRepo.CreateProfile(ctx, profile); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if input.IsDefault != nil && *input.IsDefault {
		existingAddresses, err := s.customerRepo.GetAddressesByCustomerID(ctx, profile.UserID)
		if err != nil {
			return nil, err
		}
		for _, addr := range existingAddresses {
			if addr.IsDefault {
				return nil, shared.NewConflict("a default address already exists")
			}
		}
	}

	country := "India"
	if input.Country != nil {
		country = *input.Country
	}

	address := &model.CustomerAddress{
		ID:           uuid.New(),
		CustomerID:   profile.UserID,
		AddressLine1: input.AddressLine1,
		AddressLine2: input.AddressLine2,
		City:         input.City,
		State:        input.State,
		PostalCode:   input.PostalCode,
		Country:      country,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		Label:        input.Label,
		IsDefault:    input.IsDefault != nil && *input.IsDefault,
		CreatedAt:    shared.Now(),
		UpdatedAt:    shared.Now(),
	}

	if err := s.customerRepo.CreateAddress(ctx, address); err != nil {
		return nil, err
	}

	output := address.ToOutput()
	return &output, nil
}

func (s *CustomerService) GetMyAddresses(ctx context.Context, userID uuid.UUID) ([]model.CustomerAddressOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return []model.CustomerAddressOutput{}, nil
		}
		return nil, err
	}

	addresses, err := s.customerRepo.GetAddressesByCustomerID(ctx, profile.UserID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.CustomerAddressOutput, len(addresses))
	for i, addr := range addresses {
		outputs[i] = addr.ToOutput()
	}
	return outputs, nil
}

func (s *CustomerService) GetAddressByID(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) (*model.CustomerAddressOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	address, err := s.customerRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.CustomerID != profile.UserID {
		return nil, shared.ErrNotFound
	}

	output := address.ToOutput()
	return &output, nil
}

func (s *CustomerService) UpdateAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID, input model.UpdateCustomerAddressInput) (*model.CustomerAddressOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	address, err := s.customerRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	if address.CustomerID != profile.UserID {
		return nil, shared.ErrNotFound
	}

	if input.IsDefault != nil && *input.IsDefault {
		existingAddresses, err := s.customerRepo.GetAddressesByCustomerID(ctx, profile.UserID)
		if err != nil {
			return nil, err
		}
		for _, addr := range existingAddresses {
			if addr.IsDefault && addr.ID != addressID {
				return nil, shared.NewConflict("a default address already exists")
			}
		}
	}

	updated, err := s.customerRepo.UpdateAddress(ctx, addressID, input)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *CustomerService) DeleteAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) error {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	address, err := s.customerRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address.CustomerID != profile.UserID {
		return shared.ErrNotFound
	}

	return s.customerRepo.DeleteAddress(ctx, addressID)
}

func (s *CustomerService) GetDefaultAddress(ctx context.Context, userID uuid.UUID) (*model.CustomerAddressOutput, error) {
	profile, err := s.customerRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	address, err := s.customerRepo.GetDefaultAddress(ctx, profile.UserID)
	if err != nil {
		return nil, err
	}

	output := address.ToOutput()
	return &output, nil
}
