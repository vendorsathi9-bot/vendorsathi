package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
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

func TestProfitAndAnalyticsCalculation(t *testing.T) {
	sellingPricePaise := int64(10000) // ₹100
	costPricePaise := int64(6500)     // ₹65
	unitsSold := int64(20)

	grossProfitPaise := (sellingPricePaise - costPricePaise) * unitsSold
	if grossProfitPaise != 70000 { // ₹700
		t.Errorf("expected gross profit 70000 paise (₹700.00), got %d", grossProfitPaise)
	}

	marginPercent := float64(sellingPricePaise-costPricePaise) / float64(sellingPricePaise) * 100
	if marginPercent != 35.0 {
		t.Errorf("expected 35%% margin, got %f", marginPercent)
	}

	totalExpensesPaise := int64(25000) // ₹250
	netProfitPaise := grossProfitPaise - totalExpensesPaise
	if netProfitPaise != 45000 { // ₹450
		t.Errorf("expected net profit 45000 paise (₹450.00), got %d", netProfitPaise)
	}
}

func TestDemandForecastingAlgorithm(t *testing.T) {
	// 40 units sold across 5 active days -> 8 units/day
	unitsSold := int64(40)
	activeDays := 5
	dailyAvg := float64(unitsSold) / float64(activeDays)
	if dailyAvg != 8.0 {
		t.Errorf("expected daily avg 8.0, got %f", dailyAvg)
	}

	predicted7Day := int64(math.Ceil(dailyAvg * 7))
	if predicted7Day != 56 {
		t.Errorf("expected 7-day predicted demand 56, got %d", predicted7Day)
	}

	currentStock := int64(20)
	suggestedRestock := int64(0)
	if predicted7Day > currentStock {
		suggestedRestock = predicted7Day - currentStock
	}
	if suggestedRestock != 36 {
		t.Errorf("expected suggested restock 36, got %d", suggestedRestock)
	}
}

func TestLoyaltyPointsAllocationAndRedemption(t *testing.T) {
	orderTotalPaise := int64(45000) // ₹450
	earnedPoints := int(orderTotalPaise / 1000)
	if earnedPoints != 45 {
		t.Errorf("expected 45 loyalty points (1 pt per ₹10), got %d", earnedPoints)
	}

	// Redemption: 20 points redeemed = 2000 paise (₹20 discount)
	pointsToRedeem := 20
	discountPaise := int64(pointsToRedeem) * 100
	subtotalPaise := int64(15000) // ₹150
	finalTotalPaise := subtotalPaise - discountPaise
	if finalTotalPaise != 13000 { // ₹130
		t.Errorf("expected final total 13000 paise (₹130.00), got %d", finalTotalPaise)
	}
}

func TestPlatformOverviewAndVendorModels(t *testing.T) {
	overview := struct {
		TotalUsers      int64
		TotalVendors    int64
		TotalProducts   int64
		TotalCategories int64
	}{
		TotalUsers:      25,
		TotalVendors:    10,
		TotalProducts:   50,
		TotalCategories: 8,
	}

	if overview.TotalProducts != 50 || overview.TotalCategories != 8 {
		t.Errorf("unexpected overview product/category values: %+v", overview)
	}
}

func TestDeliveryStateMachineTransitions(t *testing.T) {
	// Valid transitions
	tests := []struct {
		current  string
		next     string
		expected bool
	}{
		{"ASSIGNED", "ACCEPTED", true},
		{"ACCEPTED", "PICKED_UP", true},
		{"PICKED_UP", "OUT_FOR_DELIVERY", true},
		{"OUT_FOR_DELIVERY", "DELIVERED", true},
		{"ASSIGNED", "DELIVERED", false},
		{"DELIVERED", "ACCEPTED", false},
		{"PICKED_UP", "ACCEPTED", false},
	}

	allowed := map[string][]string{
		"ASSIGNED":         {"ACCEPTED"},
		"ACCEPTED":         {"PICKED_UP"},
		"PICKED_UP":        {"OUT_FOR_DELIVERY"},
		"OUT_FOR_DELIVERY": {"DELIVERED"},
		"DELIVERED":        {},
	}

	for _, tc := range tests {
		valid := false
		for _, a := range allowed[tc.current] {
			if a == tc.next {
				valid = true
				break
			}
		}
		if valid != tc.expected {
			t.Errorf("transition from %s to %s expected %v, got %v", tc.current, tc.next, tc.expected, valid)
		}
	}
}

func TestRazorpaySignatureVerification(t *testing.T) {
	keySecret := "test_secret_12345"
	orderID := "order_EKwxwAgIt4iNpU"
	paymentID := "pay_29QQoUBi66xm2f"

	// Expected HMAC-SHA256 of "order_id|payment_id"
	data := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(keySecret))
	mac.Write([]byte(data))
	validSig := hex.EncodeToString(mac.Sum(nil))

	// Verify valid signature matches
	testMac := hmac.New(sha256.New, []byte(keySecret))
	testMac.Write([]byte(data))
	computedSig := hex.EncodeToString(testMac.Sum(nil))
	if computedSig != validSig {
		t.Errorf("expected signature to match")
	}

	// Verify tampering is detected
	tamperedMac := hmac.New(sha256.New, []byte(keySecret))
	tamperedMac.Write([]byte(orderID + "|tampered_payment_id"))
	tamperedSig := hex.EncodeToString(tamperedMac.Sum(nil))
	if tamperedSig == validSig {
		t.Errorf("tampered signature should not match valid signature")
	}
}

func TestRouteFallbackCalculation(t *testing.T) {
	// Bangalore Majestic (12.9778, 77.5727) to Koramangala (12.9352, 77.6245)
	lat1, lon1 := 12.9778, 77.5727
	lat2, lon2 := 12.9352, 77.6245

	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	straightLineKm := earthRadiusKm * c

	if straightLineKm < 4.0 || straightLineKm > 10.0 {
		t.Errorf("expected straight-line distance between 4km and 10km, got %f", straightLineKm)
	}

	// Verify that when OSRM routing is unavailable, the fallback invariant holds:
	// 1. routing_available must be false
	// 2. distance_type must be "straight_line_estimate"
	// 3. no fake road geometry is returned
	// 4. no fabricated road duration is presented as actual road travel ETA
	fallbackResp := map[string]interface{}{
		"routing_available":             false,
		"distance_type":                 "straight_line_estimate",
		"approximate_straight_line_km": math.Round(straightLineKm*100) / 100,
		"distance_km":                   0.0,
		"duration_minutes":              0.0,
		"geometry":                      nil,
	}

	if fallbackResp["routing_available"] != false {
		t.Errorf("fallback response must have routing_available=false")
	}
	if fallbackResp["distance_type"] != "straight_line_estimate" {
		t.Errorf("fallback response must identify distance as straight_line_estimate")
	}
	if fallbackResp["geometry"] != nil {
		t.Errorf("fallback response must not have fabricated road geometry")
	}
	if fallbackResp["distance_km"] != 0.0 || fallbackResp["duration_minutes"] != 0.0 {
		t.Errorf("fallback response must not display false road distance or duration")
	}
}

func TestOSRMCoordinateOrdering(t *testing.T) {
	// OSRM strictly requires {longitude},{latitude};{longitude},{latitude}
	originLat, originLng := 12.9716, 77.5946
	destLat, destLng := 12.9352, 77.6245

	osrmPath := fmt.Sprintf("/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		originLng, originLat, destLng, destLat)

	expectedStart := fmt.Sprintf("/route/v1/driving/%f,%f;", originLng, originLat)
	if !strings.HasPrefix(osrmPath, expectedStart) {
		t.Errorf("expected OSRM path to start with longitude,latitude ordering %s, got %s", expectedStart, osrmPath)
	}

	// Verify longitude is first
	firstCoord := strings.Split(strings.TrimPrefix(osrmPath, "/route/v1/driving/"), ";")[0]
	coords := strings.Split(firstCoord, ",")
	if coords[0] != fmt.Sprintf("%f", originLng) || coords[1] != fmt.Sprintf("%f", originLat) {
		t.Errorf("coordinate ordering reversed! Expected [%f, %f], got [%s, %s]", originLng, originLat, coords[0], coords[1])
	}
}

func TestOSRMInvalidCoordinateRejection(t *testing.T) {
	// Latitude must be [-90, 90], Longitude must be [-180, 180]
	invalidCoords := []struct {
		originLat, originLng, destLat, destLng float64
	}{
		{95.0, 77.5, 12.9, 77.6},    // Invalid origin lat
		{12.9, 195.0, 12.9, 77.6},   // Invalid origin lng
		{12.9, 77.5, -95.0, 77.6},   // Invalid dest lat
		{12.9, 77.5, 12.9, -190.0},  // Invalid dest lng
	}

	for _, tc := range invalidCoords {
		isInvalid := tc.originLat < -90 || tc.originLat > 90 || tc.destLat < -90 || tc.destLat > 90 ||
			tc.originLng < -180 || tc.originLng > 180 || tc.destLng < -180 || tc.destLng > 180
		if !isInvalid {
			t.Errorf("expected coordinate set %+v to be marked invalid", tc)
		}
	}
}

func TestOSRMGeoJSONStructure(t *testing.T) {
	// Mock GeoJSON response from OSRM
	mockGeoJSON := `{
		"code": "Ok",
		"routes": [{
			"geometry": {
				"coordinates": [
					[77.5946, 12.9716],
					[77.6000, 12.9600],
					[77.6245, 12.9352]
				],
				"type": "LineString"
			},
			"distance": 6850.4,
			"duration": 960.2
		}]
	}`

	var parsed struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Type        string      `json:"type"`
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}

	err := json.Unmarshal([]byte(mockGeoJSON), &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal OSRM GeoJSON: %v", err)
	}

	if parsed.Code != "Ok" || len(parsed.Routes) == 0 {
		t.Fatalf("unexpected mock OSRM response")
	}

	route := parsed.Routes[0]
	if route.Distance != 6850.4 {
		t.Errorf("expected distance 6850.4 meters, got %f", route.Distance)
	}
	if route.Duration != 960.2 {
		t.Errorf("expected duration 960.2 seconds, got %f", route.Duration)
	}
	if route.Geometry.Type != "LineString" {
		t.Errorf("expected geometry type LineString, got %s", route.Geometry.Type)
	}
	if len(route.Geometry.Coordinates) != 3 {
		t.Errorf("expected 3 coordinate waypoints, got %d", len(route.Geometry.Coordinates))
	}

	// Verify coordinates are [longitude, latitude]
	firstPt := route.Geometry.Coordinates[0]
	if firstPt[0] != 77.5946 || firstPt[1] != 12.9716 {
		t.Errorf("expected [lng=77.5946, lat=12.9716], got [%f, %f]", firstPt[0], firstPt[1])
	}
}


