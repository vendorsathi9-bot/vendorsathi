package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/modules/delivery/model"
	"vendorsaathi/internal/shared"
)

type DeliveryRepository struct {
	db *pgxpool.Pool
}

func NewDeliveryRepository(db *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

// AssignDelivery creates or updates a delivery assignment for an order and sets orders.delivery_agent_id.
func (r *DeliveryRepository) AssignDelivery(ctx context.Context, orderID, agentUserID, assignedBy uuid.UUID) (*model.DeliveryAssignment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO delivery_assignments (order_id, agent_user_id, assigned_by, status)
		VALUES ($1, $2, $3, 'ASSIGNED')
		ON CONFLICT (order_id) DO UPDATE
		  SET agent_user_id = EXCLUDED.agent_user_id,
		      assigned_by   = EXCLUDED.assigned_by,
		      status        = 'ASSIGNED',
		      accepted_at   = NULL,
		      picked_up_at  = NULL,
		      delivered_at  = NULL,
		      assigned_at   = NOW(),
		      updated_at    = NOW()
		RETURNING id, order_id, agent_user_id, status, assigned_by, assigned_at, accepted_at, picked_up_at, delivered_at, updated_at`

	var a model.DeliveryAssignment
	err = tx.QueryRow(ctx, q, orderID, agentUserID, assignedBy).Scan(
		&a.ID, &a.OrderID, &a.AgentUserID, &a.Status, &a.AssignedBy,
		&a.AssignedAt, &a.AcceptedAt, &a.PickedUpAt, &a.DeliveredAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert delivery assignment: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE orders SET delivery_agent_id = $1, updated_at = NOW() WHERE id = $2`, agentUserID, orderID)
	if err != nil {
		return nil, fmt.Errorf("update order delivery_agent_id: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &a, nil
}

// GetAssignmentByOrderID returns the delivery assignment for the given order.
func (r *DeliveryRepository) GetAssignmentByOrderID(ctx context.Context, orderID uuid.UUID) (*model.DeliveryAssignment, error) {
	const q = `
		SELECT id, order_id, agent_user_id, status, assigned_by, assigned_at,
		       accepted_at, picked_up_at, delivered_at, updated_at
		FROM delivery_assignments WHERE order_id = $1`

	var a model.DeliveryAssignment
	err := r.db.QueryRow(ctx, q, orderID).Scan(
		&a.ID, &a.OrderID, &a.AgentUserID, &a.Status, &a.AssignedBy,
		&a.AssignedAt, &a.AcceptedAt, &a.PickedUpAt, &a.DeliveredAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get assignment: %w", err)
	}
	return &a, nil
}

// GetAssignmentsByAgent returns all assignments for the given agent, with order/vendor/customer details.
func (r *DeliveryRepository) GetAssignmentsByAgent(ctx context.Context, agentUserID uuid.UUID) ([]model.DeliveryAssignmentWithOrder, error) {
	const q = `
		SELECT
		    da.id, da.order_id, da.agent_user_id, da.status, da.assigned_by,
		    da.assigned_at, da.accepted_at, da.picked_up_at, da.delivered_at, da.updated_at,
		    COALESCE(cp.display_name, u.phone, '') AS customer_name,
		    u.phone AS customer_phone,
		    COALESCE(o.delivery_address_snapshot->>'address_line1', o.delivery_address_snapshot->>'line1', '') AS delivery_address_line1,
		    COALESCE(o.delivery_address_snapshot->>'city', '') AS delivery_city,
		    COALESCE((o.delivery_address_snapshot->>'latitude')::double precision, 0) AS delivery_lat,
		    COALESCE((o.delivery_address_snapshot->>'longitude')::double precision, 0) AS delivery_lng,
		    o.total_paise AS order_total_paise,
		    o.status AS order_status,
		    COALESCE(vp.business_name, '') AS vendor_name,
		    COALESCE(vp.phone, '') AS vendor_phone,
		    COALESCE(vl.latitude::double precision, 0) AS vendor_lat,
		    COALESCE(vl.longitude::double precision, 0) AS vendor_lng
		FROM delivery_assignments da
		JOIN orders o ON o.id = da.order_id
		JOIN users u ON u.id = o.customer_id
		LEFT JOIN customer_profiles cp ON cp.user_id = o.customer_id
		JOIN vendor_profiles vp ON vp.id = o.vendor_id
		LEFT JOIN vendor_locations vl ON vl.vendor_id = vp.id AND vl.is_primary = TRUE
		WHERE da.agent_user_id = $1
		ORDER BY da.assigned_at DESC`

	rows, err := r.db.Query(ctx, q, agentUserID)
	if err != nil {
		return nil, fmt.Errorf("list agent assignments: %w", err)
	}
	defer rows.Close()

	var results []model.DeliveryAssignmentWithOrder
	for rows.Next() {
		var a model.DeliveryAssignmentWithOrder
		err := rows.Scan(
			&a.ID, &a.OrderID, &a.AgentUserID, &a.Status, &a.AssignedBy,
			&a.AssignedAt, &a.AcceptedAt, &a.PickedUpAt, &a.DeliveredAt, &a.UpdatedAt,
			&a.CustomerName, &a.CustomerPhone,
			&a.DeliveryAddressLine1, &a.DeliveryCity, &a.DeliveryLat, &a.DeliveryLng,
			&a.OrderTotalPaise, &a.OrderStatus,
			&a.VendorName, &a.VendorPhone, &a.VendorLat, &a.VendorLng,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent assignment: %w", err)
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

// UpdateAssignmentStatus updates the delivery status and also syncs order status if applicable.
func (r *DeliveryRepository) UpdateAssignmentStatus(ctx context.Context, orderID uuid.UUID, newStatus model.DeliveryStatus) (*model.DeliveryAssignment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	var q string
	var args []interface{}

	switch newStatus {
	case model.DeliveryStatusAccepted:
		q = `UPDATE delivery_assignments SET status=$1, accepted_at=$2, updated_at=$2 WHERE order_id=$3
		     RETURNING id, order_id, agent_user_id, status, assigned_by, assigned_at, accepted_at, picked_up_at, delivered_at, updated_at`
		args = []interface{}{newStatus, now, orderID}
	case model.DeliveryStatusPickedUp:
		q = `UPDATE delivery_assignments SET status=$1, picked_up_at=$2, updated_at=$2 WHERE order_id=$3
		     RETURNING id, order_id, agent_user_id, status, assigned_by, assigned_at, accepted_at, picked_up_at, delivered_at, updated_at`
		args = []interface{}{newStatus, now, orderID}
	case model.DeliveryStatusOutForDelivery:
		q = `UPDATE delivery_assignments SET status=$1, updated_at=$2 WHERE order_id=$3
		     RETURNING id, order_id, agent_user_id, status, assigned_by, assigned_at, accepted_at, picked_up_at, delivered_at, updated_at`
		args = []interface{}{newStatus, now, orderID}
	case model.DeliveryStatusDelivered:
		q = `UPDATE delivery_assignments SET status=$1, delivered_at=$2, updated_at=$2 WHERE order_id=$3
		     RETURNING id, order_id, agent_user_id, status, assigned_by, assigned_at, accepted_at, picked_up_at, delivered_at, updated_at`
		args = []interface{}{newStatus, now, orderID}
	default:
		return nil, fmt.Errorf("unknown delivery status: %s", newStatus)
	}

	var a model.DeliveryAssignment
	err = tx.QueryRow(ctx, q, args...).Scan(
		&a.ID, &a.OrderID, &a.AgentUserID, &a.Status, &a.AssignedBy,
		&a.AssignedAt, &a.AcceptedAt, &a.PickedUpAt, &a.DeliveredAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("update delivery status: %w", err)
	}

	// Sync order status when out for delivery or delivered
	if newStatus == model.DeliveryStatusOutForDelivery {
		_, _ = tx.Exec(ctx, `UPDATE orders SET status = 'OUT_FOR_DELIVERY', updated_at = NOW() WHERE id = $1 AND status != 'OUT_FOR_DELIVERY'`, orderID)
	} else if newStatus == model.DeliveryStatusDelivered {
		_, _ = tx.Exec(ctx, `UPDATE orders SET status = 'COMPLETED', updated_at = NOW() WHERE id = $1 AND status != 'COMPLETED'`, orderID)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &a, nil
}

// UpsertLocation upserts the agent's GPS coordinates for an order.
func (r *DeliveryRepository) UpsertLocation(ctx context.Context, orderID, agentUserID uuid.UUID, lat, lng float64) (*model.DeliveryTracking, error) {
	const q = `
		INSERT INTO delivery_tracking (order_id, agent_user_id, latitude, longitude, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (order_id) DO UPDATE
		  SET latitude     = EXCLUDED.latitude,
		      longitude    = EXCLUDED.longitude,
		      agent_user_id = EXCLUDED.agent_user_id,
		      updated_at   = NOW()
		RETURNING id, order_id, agent_user_id, latitude, longitude, updated_at`

	var t model.DeliveryTracking
	err := r.db.QueryRow(ctx, q, orderID, agentUserID, lat, lng).Scan(
		&t.ID, &t.OrderID, &t.AgentUserID, &t.Latitude, &t.Longitude, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert tracking: %w", err)
	}
	return &t, nil
}

// GetLocation returns the latest tracked location for an order.
func (r *DeliveryRepository) GetLocation(ctx context.Context, orderID uuid.UUID) (*model.DeliveryTracking, error) {
	const q = `
		SELECT id, order_id, agent_user_id, latitude, longitude, updated_at
		FROM delivery_tracking WHERE order_id = $1`

	var t model.DeliveryTracking
	err := r.db.QueryRow(ctx, q, orderID).Scan(
		&t.ID, &t.OrderID, &t.AgentUserID, &t.Latitude, &t.Longitude, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get tracking: %w", err)
	}
	return &t, nil
}

// ListAllAssignments returns all assignments for admin view.
func (r *DeliveryRepository) ListAllAssignments(ctx context.Context) ([]model.DeliveryAssignmentWithOrder, error) {
	const q = `
		SELECT
		    da.id, da.order_id, da.agent_user_id, da.status, da.assigned_by,
		    da.assigned_at, da.accepted_at, da.picked_up_at, da.delivered_at, da.updated_at,
		    COALESCE(cp.display_name, u.phone, '') AS customer_name,
		    u.phone AS customer_phone,
		    COALESCE(o.delivery_address_snapshot->>'address_line1', o.delivery_address_snapshot->>'line1', '') AS delivery_address_line1,
		    COALESCE(o.delivery_address_snapshot->>'city', '') AS delivery_city,
		    COALESCE((o.delivery_address_snapshot->>'latitude')::double precision, 0) AS delivery_lat,
		    COALESCE((o.delivery_address_snapshot->>'longitude')::double precision, 0) AS delivery_lng,
		    o.total_paise AS order_total_paise,
		    o.status AS order_status,
		    COALESCE(vp.business_name, '') AS vendor_name,
		    COALESCE(vp.phone, '') AS vendor_phone,
		    COALESCE(vl.latitude::double precision, 0) AS vendor_lat,
		    COALESCE(vl.longitude::double precision, 0) AS vendor_lng
		FROM delivery_assignments da
		JOIN orders o ON o.id = da.order_id
		JOIN users u ON u.id = o.customer_id
		LEFT JOIN customer_profiles cp ON cp.user_id = o.customer_id
		JOIN vendor_profiles vp ON vp.id = o.vendor_id
		LEFT JOIN vendor_locations vl ON vl.vendor_id = vp.id AND vl.is_primary = TRUE
		ORDER BY da.assigned_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list all assignments: %w", err)
	}
	defer rows.Close()

	var results []model.DeliveryAssignmentWithOrder
	for rows.Next() {
		var a model.DeliveryAssignmentWithOrder
		err := rows.Scan(
			&a.ID, &a.OrderID, &a.AgentUserID, &a.Status, &a.AssignedBy,
			&a.AssignedAt, &a.AcceptedAt, &a.PickedUpAt, &a.DeliveredAt, &a.UpdatedAt,
			&a.CustomerName, &a.CustomerPhone,
			&a.DeliveryAddressLine1, &a.DeliveryCity, &a.DeliveryLat, &a.DeliveryLng,
			&a.OrderTotalPaise, &a.OrderStatus,
			&a.VendorName, &a.VendorPhone, &a.VendorLat, &a.VendorLng,
		)
		if err != nil {
			return nil, fmt.Errorf("scan assignment: %w", err)
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

// GetOrderMeta returns customer_id (user ID) and vendor's owner_user_id for auth checks.
func (r *DeliveryRepository) GetOrderMeta(ctx context.Context, orderID uuid.UUID) (customerID, vendorOwnerID uuid.UUID, err error) {
	const q = `
		SELECT o.customer_id, vp.owner_user_id
		FROM orders o
		JOIN vendor_profiles vp ON vp.id = o.vendor_id
		WHERE o.id = $1`
	err = r.db.QueryRow(ctx, q, orderID).Scan(&customerID, &vendorOwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, uuid.Nil, shared.ErrNotFound
		}
		return uuid.Nil, uuid.Nil, fmt.Errorf("get order meta: %w", err)
	}
	return
}
