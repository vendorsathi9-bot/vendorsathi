package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/vendors/model"
	"vendorsaathi/internal/shared"
)

type VendorRepository struct {
	pool *pgxpool.Pool
}

func NewVendorRepository(pool *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{pool: pool}
}

func (r *VendorRepository) CreateProfile(ctx context.Context, profile *model.VendorProfile) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vendor_profiles (id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, profile.ID, profile.OwnerUserID, profile.BusinessName, profile.Description, profile.Phone, profile.Status, profile.AvailabilityStatus, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shared.NewConflict("vendor profile already exists for this user")
		}
		return err
	}
	return nil
}

func (r *VendorRepository) GetProfileByID(ctx context.Context, id uuid.UUID) (*model.VendorProfile, error) {
	var profile model.VendorProfile
	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at
		FROM vendor_profiles WHERE id = $1
	`, id).Scan(&profile.ID, &profile.OwnerUserID, &profile.BusinessName, &profile.Description, &profile.Phone, &profile.Status, &profile.AvailabilityStatus, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *VendorRepository) GetProfileByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (*model.VendorProfile, error) {
	var profile model.VendorProfile
	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at
		FROM vendor_profiles WHERE owner_user_id = $1
	`, ownerUserID).Scan(&profile.ID, &profile.OwnerUserID, &profile.BusinessName, &profile.Description, &profile.Phone, &profile.Status, &profile.AvailabilityStatus, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *VendorRepository) UpdateProfile(ctx context.Context, id uuid.UUID, input model.UpdateVendorProfileInput) (*model.VendorProfile, error) {
	query := `
		UPDATE vendor_profiles
		SET business_name = COALESCE(NULLIF($1, ''), business_name),
		    description = COALESCE($2, description),
		    phone = COALESCE(NULLIF($3, ''), phone),
		    availability_status = COALESCE($4, availability_status),
		    updated_at = NOW()
		WHERE id = $5
		RETURNING id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at
	`
	var profile model.VendorProfile
	err := r.pool.QueryRow(ctx, query, input.BusinessName, input.Description, input.Phone, input.AvailabilityStatus, id).Scan(
		&profile.ID, &profile.OwnerUserID, &profile.BusinessName, &profile.Description, &profile.Phone, &profile.Status, &profile.AvailabilityStatus, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *VendorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.VendorStatus) (*model.VendorProfile, error) {
	var profile model.VendorProfile
	err := r.pool.QueryRow(ctx, `
		UPDATE vendor_profiles SET status = $1, updated_at = NOW() WHERE id = $2
		RETURNING id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at
	`, status, id).Scan(&profile.ID, &profile.OwnerUserID, &profile.BusinessName, &profile.Description, &profile.Phone, &profile.Status, &profile.AvailabilityStatus, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *VendorRepository) CreateLocation(ctx context.Context, location *model.VendorLocation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO vendor_locations (id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, location.ID, location.VendorID, location.AddressLine1, location.AddressLine2, location.City, location.State, location.PostalCode, location.Country, location.Latitude, location.Longitude, location.IsPrimary, location.UpdatedAt)
	return err
}

func (r *VendorRepository) GetLocationByID(ctx context.Context, id uuid.UUID) (*model.VendorLocation, error) {
	var location model.VendorLocation
	err := r.pool.QueryRow(ctx, `
		SELECT id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at
		FROM vendor_locations WHERE id = $1
	`, id).Scan(&location.ID, &location.VendorID, &location.AddressLine1, &location.AddressLine2, &location.City, &location.State, &location.PostalCode, &location.Country, &location.Latitude, &location.Longitude, &location.IsPrimary, &location.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

func (r *VendorRepository) GetLocationsByVendorID(ctx context.Context, vendorID uuid.UUID) ([]model.VendorLocation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at
		FROM vendor_locations WHERE vendor_id = $1 ORDER BY is_primary DESC, updated_at DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.VendorLocation
	for rows.Next() {
		var location model.VendorLocation
		if err := rows.Scan(&location.ID, &location.VendorID, &location.AddressLine1, &location.AddressLine2, &location.City, &location.State, &location.PostalCode, &location.Country, &location.Latitude, &location.Longitude, &location.IsPrimary, &location.UpdatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return locations, rows.Err()
}

func (r *VendorRepository) GetPrimaryLocation(ctx context.Context, vendorID uuid.UUID) (*model.VendorLocation, error) {
	var location model.VendorLocation
	err := r.pool.QueryRow(ctx, `
		SELECT id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at
		FROM vendor_locations WHERE vendor_id = $1 AND is_primary = TRUE
	`, vendorID).Scan(&location.ID, &location.VendorID, &location.AddressLine1, &location.AddressLine2, &location.City, &location.State, &location.PostalCode, &location.Country, &location.Latitude, &location.Longitude, &location.IsPrimary, &location.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

func (r *VendorRepository) UpdateLocation(ctx context.Context, id uuid.UUID, input model.UpdateVendorLocationInput) (*model.VendorLocation, error) {
	query := `
		UPDATE vendor_locations
		SET address_line1 = COALESCE(NULLIF($1, ''), address_line1),
		    address_line2 = COALESCE($2, address_line2),
		    city = COALESCE(NULLIF($3, ''), city),
		    state = COALESCE(NULLIF($4, ''), state),
		    postal_code = COALESCE($5, postal_code),
		    country = COALESCE(NULLIF($6, ''), country),
		    latitude = COALESCE($7, latitude),
		    longitude = COALESCE($8, longitude),
		    is_primary = COALESCE($9, is_primary),
		    updated_at = NOW()
		WHERE id = $10
		RETURNING id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at
	`
	var location model.VendorLocation
	err := r.pool.QueryRow(ctx, query,
		input.AddressLine1, input.AddressLine2, input.City, input.State,
		input.PostalCode, input.Country, input.Latitude, input.Longitude,
		input.IsPrimary, id).Scan(
		&location.ID, &location.VendorID, &location.AddressLine1, &location.AddressLine2,
		&location.City, &location.State, &location.PostalCode, &location.Country,
		&location.Latitude, &location.Longitude, &location.IsPrimary, &location.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

func (r *VendorRepository) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM vendor_locations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *VendorRepository) SetBusinessHours(ctx context.Context, vendorID uuid.UUID, hours []model.BusinessHourInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM business_hours WHERE vendor_id = $1`, vendorID)
	if err != nil {
		return err
	}

	for _, h := range hours {
		openTime := h.OpenTime
		closeTime := h.CloseTime
		isClosed := false
		if h.IsClosed != nil {
			isClosed = *h.IsClosed
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO business_hours (vendor_id, weekday, open_time, close_time, is_closed)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (vendor_id, weekday) DO UPDATE SET
				open_time = EXCLUDED.open_time,
				close_time = EXCLUDED.close_time,
				is_closed = EXCLUDED.is_closed
		`, vendorID, h.Weekday, openTime, closeTime, isClosed)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *VendorRepository) GetBusinessHours(ctx context.Context, vendorID uuid.UUID) ([]model.BusinessHour, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, vendor_id, weekday, open_time, close_time, is_closed
		FROM business_hours WHERE vendor_id = $1 ORDER BY weekday
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hours []model.BusinessHour
	for rows.Next() {
		var hour model.BusinessHour
		if err := rows.Scan(&hour.ID, &hour.VendorID, &hour.Weekday, &hour.OpenTime, &hour.CloseTime, &hour.IsClosed); err != nil {
			return nil, err
		}
		hours = append(hours, hour)
	}
	return hours, rows.Err()
}
