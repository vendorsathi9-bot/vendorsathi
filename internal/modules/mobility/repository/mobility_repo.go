package repository

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/mobility/model"
)

type MobilityRepository struct {
	pool *pgxpool.Pool
}

func NewMobilityRepository(pool *pgxpool.Pool) *MobilityRepository {
	return &MobilityRepository{pool: pool}
}

// FindNearbyVendors queries active vendors within radiusKm using bounding box + Haversine formula
func (r *MobilityRepository) FindNearbyVendors(ctx context.Context, lat, lng, radiusKm float64) ([]model.NearbyVendor, error) {
	// 1 degree latitude ~= 111 km
	latDelta := radiusKm / 111.0
	minLat := lat - latDelta
	maxLat := lat + latDelta

	// 1 degree longitude ~= 111 * cos(radians(lat)) km
	radLat := lat * math.Pi / 180.0
	cosLat := math.Cos(radLat)
	if cosLat < 0.0001 {
		cosLat = 0.0001
	}
	lngDelta := radiusKm / (111.0 * cosLat)
	minLng := lng - lngDelta
	maxLng := lng + lngDelta

	query := `
		SELECT 
			vp.id,
			vp.business_name,
			vp.description,
			vp.phone,
			vp.status,
			vp.availability_status,
			vl.latitude,
			vl.longitude,
			vl.address_line1,
			vl.city,
			(6371 * acos(
				LEAST(1.0, GREATEST(-1.0,
					cos(radians($1)) * cos(radians(vl.latitude)) * cos(radians(vl.longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(vl.latitude))
				))
			)) AS distance_km
		FROM vendor_profiles vp
		JOIN vendor_locations vl ON vp.id = vl.vendor_id
		WHERE vp.status = 'ACTIVE'
		  AND vl.latitude BETWEEN $3 AND $4
		  AND vl.longitude BETWEEN $5 AND $6
		  AND (6371 * acos(
				LEAST(1.0, GREATEST(-1.0,
					cos(radians($1)) * cos(radians(vl.latitude)) * cos(radians(vl.longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(vl.latitude))
				))
		  )) <= $7
		ORDER BY distance_km ASC
		LIMIT 50;
	`

	rows, err := r.pool.Query(ctx, query, lat, lng, minLat, maxLat, minLng, maxLng, radiusKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.NearbyVendor
	for rows.Next() {
		var v model.NearbyVendor
		if err := rows.Scan(
			&v.VendorID,
			&v.BusinessName,
			&v.Description,
			&v.Phone,
			&v.Status,
			&v.AvailabilityStatus,
			&v.Latitude,
			&v.Longitude,
			&v.AddressLine1,
			&v.City,
			&v.DistanceKm,
		); err != nil {
			return nil, err
		}
		results = append(results, v)
	}

	return results, rows.Err()
}
