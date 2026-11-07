package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/customers/model"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/shared"
)

type CustomerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

func (r *CustomerRepository) db(ctx context.Context) database.DBTX {
	return database.DBFromContext(ctx, r.pool)
}

func (r *CustomerRepository) CreateProfile(ctx context.Context, profile *model.CustomerProfile) error {
	_, err := r.db(ctx).Exec(ctx, `
		INSERT INTO customer_profiles (user_id, display_name, contact_preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, profile.UserID, profile.DisplayName, profile.ContactPreferences, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *CustomerRepository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*model.CustomerProfile, error) {
	var profile model.CustomerProfile
	err := r.db(ctx).QueryRow(ctx, `
		SELECT user_id, display_name, contact_preferences, created_at, updated_at
		FROM customer_profiles WHERE user_id = $1
	`, userID).Scan(&profile.UserID, &profile.DisplayName, &profile.ContactPreferences, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *CustomerRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, input model.UpdateCustomerProfileInput) (*model.CustomerProfile, error) {
	query := `
		UPDATE customer_profiles
		SET display_name = COALESCE($1, display_name),
		    contact_preferences = COALESCE($2, contact_preferences),
		    updated_at = NOW()
		WHERE user_id = $3
		RETURNING user_id, display_name, contact_preferences, created_at, updated_at
	`
	var profile model.CustomerProfile
	err := r.db(ctx).QueryRow(ctx, query, input.DisplayName, input.ContactPreferences, userID).Scan(
		&profile.UserID, &profile.DisplayName, &profile.ContactPreferences, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *CustomerRepository) CreateAddress(ctx context.Context, address *model.CustomerAddress) error {
	_, err := r.db(ctx).Exec(ctx, `
		INSERT INTO customer_addresses (id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, address.ID, address.CustomerID, address.AddressLine1, address.AddressLine2, address.City, address.State, address.PostalCode, address.Country, address.Latitude, address.Longitude, address.Label, address.IsDefault, address.CreatedAt, address.UpdatedAt)
	return err
}

func (r *CustomerRepository) GetAddressByID(ctx context.Context, id uuid.UUID) (*model.CustomerAddress, error) {
	var address model.CustomerAddress
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at
		FROM customer_addresses WHERE id = $1
	`, id).Scan(&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.PostalCode, &address.Country, &address.Latitude, &address.Longitude, &address.Label, &address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

func (r *CustomerRepository) GetAddressesByCustomerID(ctx context.Context, customerID uuid.UUID) ([]model.CustomerAddress, error) {
	rows, err := r.db(ctx).Query(ctx, `
		SELECT id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at
		FROM customer_addresses WHERE customer_id = $1 ORDER BY is_default DESC, updated_at DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []model.CustomerAddress
	for rows.Next() {
		var address model.CustomerAddress
		if err := rows.Scan(&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.PostalCode, &address.Country, &address.Latitude, &address.Longitude, &address.Label, &address.IsDefault, &address.CreatedAt, &address.UpdatedAt); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}
	return addresses, rows.Err()
}

func (r *CustomerRepository) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*model.CustomerAddress, error) {
	var address model.CustomerAddress
	err := r.db(ctx).QueryRow(ctx, `
		SELECT id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at
		FROM customer_addresses WHERE customer_id = $1 AND is_default = TRUE
	`, customerID).Scan(&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.PostalCode, &address.Country, &address.Latitude, &address.Longitude, &address.Label, &address.IsDefault, &address.CreatedAt, &address.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

func (r *CustomerRepository) UpdateAddress(ctx context.Context, id uuid.UUID, input model.UpdateCustomerAddressInput) (*model.CustomerAddress, error) {
	query := `
		UPDATE customer_addresses
		SET address_line1 = COALESCE(NULLIF($1, ''), address_line1),
		    address_line2 = COALESCE($2, address_line2),
		    city = COALESCE(NULLIF($3, ''), city),
		    state = COALESCE(NULLIF($4, ''), state),
		    postal_code = COALESCE($5, postal_code),
		    country = COALESCE(NULLIF($6, ''), country),
		    latitude = COALESCE($7, latitude),
		    longitude = COALESCE($8, longitude),
		    label = COALESCE($9, label),
		    is_default = COALESCE($10, is_default),
		    updated_at = NOW()
		WHERE id = $11
		RETURNING id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at
	`
	var address model.CustomerAddress
	err := r.db(ctx).QueryRow(ctx, query,
		input.AddressLine1, input.AddressLine2, input.City, input.State,
		input.PostalCode, input.Country, input.Latitude, input.Longitude,
		input.Label, input.IsDefault, id).Scan(
		&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2,
		&address.City, &address.State, &address.PostalCode, &address.Country,
		&address.Latitude, &address.Longitude, &address.Label, &address.IsDefault,
		&address.CreatedAt, &address.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

func (r *CustomerRepository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	cmdTag, err := r.db(ctx).Exec(ctx, `DELETE FROM customer_addresses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}
