package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/vendors/model"
	"vendorsaathi/internal/modules/vendors/repository"
	"vendorsaathi/internal/shared"
)

type VendorService struct {
	vendorRepo *repository.VendorRepository
}

func NewVendorService(vendorRepo *repository.VendorRepository) *VendorService {
	return &VendorService{vendorRepo: vendorRepo}
}

func (s *VendorService) CreateProfile(ctx context.Context, ownerUserID uuid.UUID, input model.CreateVendorProfileInput) (*model.VendorProfileOutput, error) {
	existing, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err == nil && existing != nil {
		return nil, shared.NewConflict("vendor profile already exists for this user")
	}
	if !errors.Is(err, shared.ErrNotFound) && err != nil {
		return nil, err
	}

	profile := &model.VendorProfile{
		ID:                 uuid.New(),
		OwnerUserID:        ownerUserID,
		BusinessName:       input.BusinessName,
		Description:        input.Description,
		Phone:              input.Phone,
		Status:             model.VendorStatusPending,
		AvailabilityStatus: model.AvailabilityClosed,
		CreatedAt:          shared.Now(),
		UpdatedAt:          shared.Now(),
	}

	if err := s.vendorRepo.CreateProfile(ctx, profile); err != nil {
		return nil, err
	}

	output := profile.ToOutput()
	return &output, nil
}

func (s *VendorService) GetMyProfile(ctx context.Context, ownerUserID uuid.UUID) (*model.VendorProfileOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	output := profile.ToOutput()
	return &output, nil
}

func (s *VendorService) GetProfileByID(ctx context.Context, id uuid.UUID) (*model.VendorProfileOutput, error) {
	profile, err := s.vendorRepo.GetProfileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	output := profile.ToOutput()
	return &output, nil
}

func (s *VendorService) UpdateMyProfile(ctx context.Context, ownerUserID uuid.UUID, input model.UpdateVendorProfileInput) (*model.VendorProfileOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	if input.Phone != nil && *input.Phone != profile.Phone {
		return nil, shared.NewValidationError("phone", "phone cannot be changed after creation")
	}

	updated, err := s.vendorRepo.UpdateProfile(ctx, profile.ID, input)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *VendorService) UpdateProfileStatus(ctx context.Context, actorID uuid.UUID, isAdmin bool, id uuid.UUID, input model.UpdateVendorStatusInput) (*model.VendorProfileOutput, error) {
	if !isAdmin {
		return nil, shared.NewForbidden("update vendor status")
	}

	updated, err := s.vendorRepo.UpdateStatus(ctx, id, input.Status)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *VendorService) CreateLocation(ctx context.Context, ownerUserID uuid.UUID, input model.CreateVendorLocationInput) (*model.VendorLocationOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	if input.IsPrimary != nil && *input.IsPrimary {
		existingLocations, err := s.vendorRepo.GetLocationsByVendorID(ctx, profile.ID)
		if err != nil {
			return nil, err
		}
		for _, loc := range existingLocations {
			if loc.IsPrimary {
				return nil, shared.NewConflict("a primary location already exists")
			}
		}
	}

	country := "India"
	if input.Country != nil {
		country = *input.Country
	}

	location := &model.VendorLocation{
		ID:           uuid.New(),
		VendorID:     profile.ID,
		AddressLine1: input.AddressLine1,
		AddressLine2: input.AddressLine2,
		City:         input.City,
		State:        input.State,
		PostalCode:   input.PostalCode,
		Country:      country,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		IsPrimary:    input.IsPrimary != nil && *input.IsPrimary,
		UpdatedAt:    shared.Now(),
	}

	if err := s.vendorRepo.CreateLocation(ctx, location); err != nil {
		return nil, err
	}

	output := location.ToOutput()
	return &output, nil
}

func (s *VendorService) GetMyLocations(ctx context.Context, ownerUserID uuid.UUID) ([]model.VendorLocationOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	locations, err := s.vendorRepo.GetLocationsByVendorID(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.VendorLocationOutput, len(locations))
	for i, loc := range locations {
		outputs[i] = loc.ToOutput()
	}
	return outputs, nil
}

func (s *VendorService) GetLocationByID(ctx context.Context, ownerUserID uuid.UUID, locationID uuid.UUID) (*model.VendorLocationOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	location, err := s.vendorRepo.GetLocationByID(ctx, locationID)
	if err != nil {
		return nil, err
	}

	if location.VendorID != profile.ID {
		return nil, shared.ErrNotFound
	}

	output := location.ToOutput()
	return &output, nil
}

func (s *VendorService) UpdateLocation(ctx context.Context, ownerUserID uuid.UUID, locationID uuid.UUID, input model.UpdateVendorLocationInput) (*model.VendorLocationOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	location, err := s.vendorRepo.GetLocationByID(ctx, locationID)
	if err != nil {
		return nil, err
	}

	if location.VendorID != profile.ID {
		return nil, shared.ErrNotFound
	}

	if input.IsPrimary != nil && *input.IsPrimary {
		existingLocations, err := s.vendorRepo.GetLocationsByVendorID(ctx, profile.ID)
		if err != nil {
			return nil, err
		}
		for _, loc := range existingLocations {
			if loc.IsPrimary && loc.ID != locationID {
				return nil, shared.NewConflict("a primary location already exists")
			}
		}
	}

	updated, err := s.vendorRepo.UpdateLocation(ctx, locationID, input)
	if err != nil {
		return nil, err
	}

	output := updated.ToOutput()
	return &output, nil
}

func (s *VendorService) DeleteLocation(ctx context.Context, ownerUserID uuid.UUID, locationID uuid.UUID) error {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return err
	}

	location, err := s.vendorRepo.GetLocationByID(ctx, locationID)
	if err != nil {
		return err
	}

	if location.VendorID != profile.ID {
		return shared.ErrNotFound
	}

	return s.vendorRepo.DeleteLocation(ctx, locationID)
}

func (s *VendorService) SetBusinessHours(ctx context.Context, ownerUserID uuid.UUID, input model.SetBusinessHoursInput) ([]model.BusinessHourOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	if err := s.vendorRepo.SetBusinessHours(ctx, profile.ID, input.Hours); err != nil {
		return nil, err
	}

	hours, err := s.vendorRepo.GetBusinessHours(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.BusinessHourOutput, len(hours))
	for i, h := range hours {
		outputs[i] = h.ToOutput()
	}
	return outputs, nil
}

func (s *VendorService) GetBusinessHours(ctx context.Context, ownerUserID uuid.UUID) ([]model.BusinessHourOutput, error) {
	profile, err := s.vendorRepo.GetProfileByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	hours, err := s.vendorRepo.GetBusinessHours(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	outputs := make([]model.BusinessHourOutput, len(hours))
	for i, h := range hours {
		outputs[i] = h.ToOutput()
	}
	return outputs, nil
}

func (s *VendorService) GetVendorDetail(ctx context.Context, vendorID uuid.UUID) (*model.VendorDetailOutput, error) {
	profile, err := s.vendorRepo.GetProfileByID(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	locations, err := s.vendorRepo.GetLocationsByVendorID(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	hours, err := s.vendorRepo.GetBusinessHours(ctx, vendorID)
	if err != nil {
		return nil, err
	}

	locOutputs := make([]model.VendorLocationOutput, len(locations))
	for i, loc := range locations {
		locOutputs[i] = loc.ToOutput()
	}

	hourOutputs := make([]model.BusinessHourOutput, len(hours))
	for i, h := range hours {
		hourOutputs[i] = h.ToOutput()
	}

	return &model.VendorDetailOutput{
		Profile:       profile.ToOutput(),
		Locations:     locOutputs,
		BusinessHours: hourOutputs,
	}, nil
}

func (s *VendorService) ListVendors(ctx context.Context) ([]model.VendorListItem, error) {
	return s.vendorRepo.ListVendors(ctx)
}
