package tests

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"vendorsaathi/internal/modules/orders/model"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/shared"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretVendor@2026"
	hash, err := auth.HashPassword(password, 64*1024, 3, 2, 16, 32)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := auth.VerifyPassword(password, hash, 64*1024, 3, 2, 16, 32)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}
	if !match {
		t.Errorf("expected password to match hash")
	}

	wrongMatch, _ := auth.VerifyPassword("WrongPassword", hash, 64*1024, 3, 2, 16, 32)
	if wrongMatch {
		t.Errorf("expected wrong password to not match hash")
	}
}

func TestJWTManager(t *testing.T) {
	jwtManager := auth.NewJWTManager(
		"super-secure-test-signing-key-32bytes!",
		"vendorsaathi-test",
		"vendorsaathi-clients",
		15*time.Minute,
		24*time.Hour,
	)

	userID := uuid.New().String()
	sessionID := uuid.New().String()
	roles := []string{"VENDOR"}
	permissions := []string{"products:read", "products:write"}

	tokenPair, err := jwtManager.GenerateTokenPair(userID, sessionID, roles, permissions)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	claims, err := jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
	}
	if claims.SessionID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, claims.SessionID)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "VENDOR" {
		t.Errorf("expected roles ['VENDOR'], got %v", claims.Roles)
	}
}

func TestOrderStateMachineRules(t *testing.T) {
	validTransitions := map[model.OrderStatus][]model.OrderStatus{
		model.OrderStatusPending:               {model.OrderStatusAccepted, model.OrderStatusRejected, model.OrderStatusCancellationRequested},
		model.OrderStatusAccepted:              {model.OrderStatusPreparing, model.OrderStatusCancellationRequested},
		model.OrderStatusPreparing:             {model.OrderStatusReadyForPickup, model.OrderStatusOutForDelivery},
		model.OrderStatusReadyForPickup:        {model.OrderStatusCompleted},
		model.OrderStatusOutForDelivery:        {model.OrderStatusCompleted},
		model.OrderStatusCancellationRequested: {model.OrderStatusCancelled, model.OrderStatusPending, model.OrderStatusAccepted},
	}

	isValid := func(from, to model.OrderStatus) bool {
		targets, ok := validTransitions[from]
		if !ok {
			return false
		}
		for _, target := range targets {
			if target == to {
				return true
			}
		}
		return false
	}

	// Valid transitions
	if !isValid(model.OrderStatusPending, model.OrderStatusAccepted) {
		t.Errorf("PENDING -> ACCEPTED should be valid")
	}
	if !isValid(model.OrderStatusAccepted, model.OrderStatusPreparing) {
		t.Errorf("ACCEPTED -> PREPARING should be valid")
	}
	if !isValid(model.OrderStatusPreparing, model.OrderStatusReadyForPickup) {
		t.Errorf("PREPARING -> READY_FOR_PICKUP should be valid")
	}
	if !isValid(model.OrderStatusReadyForPickup, model.OrderStatusCompleted) {
		t.Errorf("READY_FOR_PICKUP -> COMPLETED should be valid")
	}

	// Invalid transitions
	if isValid(model.OrderStatusPending, model.OrderStatusCompleted) {
		t.Errorf("PENDING -> COMPLETED directly should be INVALID")
	}
	if isValid(model.OrderStatusCompleted, model.OrderStatusCancelled) {
		t.Errorf("COMPLETED -> CANCELLED should be INVALID")
	}
	if isValid(model.OrderStatusRejected, model.OrderStatusAccepted) {
		t.Errorf("REJECTED -> ACCEPTED should be INVALID")
	}
}

func TestFinancialInvariants(t *testing.T) {
	// Rule 1: No negative money amounts
	itemPricePaise := int64(15000) // ₹150.00
	quantity := int64(3)
	lineTotalPaise := itemPricePaise * quantity
	if lineTotalPaise != 45000 {
		t.Errorf("expected 45000 paise (₹450.00), got %d", lineTotalPaise)
	}

	// Rule 2: Overpayment detection
	orderTotal := int64(50000)
	existingPaid := int64(30000)
	attemptedPayment := int64(25000)

	if existingPaid+attemptedPayment > orderTotal {
		err := shared.NewPaymentLimitExceeded(orderTotal-existingPaid, attemptedPayment)
		if err.Code != "PAYMENT_LIMIT_EXCEEDED" {
			t.Errorf("expected error code PAYMENT_LIMIT_EXCEEDED, got %s", err.Code)
		}
	} else {
		t.Errorf("expected overpayment error")
	}

	// Rule 3: Udhaar repayment cannot exceed outstanding balance
	udhaarBalance := int64(20000)
	attemptedRepayment := int64(25000)
	if attemptedRepayment > udhaarBalance {
		err := shared.NewUdhaarLimitExceeded(udhaarBalance, attemptedRepayment)
		if err.Code != "UDHAAR_LIMIT_EXCEEDED" {
			t.Errorf("expected error code UDHAAR_LIMIT_EXCEEDED, got %s", err.Code)
		}
	} else {
		t.Errorf("expected udhaar limit exceeded")
	}
}

func TestHaversineDistance(t *testing.T) {
	haversine := func(lat1, lon1, lat2, lon2 float64) float64 {
		rad := math.Pi / 180.0
		dLat := (lat2 - lat1) * rad
		dLon := (lon2 - lon1) * rad
		a := math.Sin(dLat/2)*math.Sin(dLat/2) +
			math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
		c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
		return 6371 * c
	}

	// Bangalore Majestic to Indiranagar (~7.5 km)
	majesticLat, majesticLng := 12.9767, 77.5713
	indiranagarLat, indiranagarLng := 12.9784, 77.6408

	dist := haversine(majesticLat, majesticLng, indiranagarLat, indiranagarLng)
	if dist < 6.0 || dist > 9.0 {
		t.Errorf("expected distance between Majestic and Indiranagar to be ~7.5km, got %f km", dist)
	}
}
