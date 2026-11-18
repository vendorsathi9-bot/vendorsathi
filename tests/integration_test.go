package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	authmodel "vendorsaathi/internal/modules/auth/model"
	authrepository "vendorsaathi/internal/modules/auth/repository"
	authservice "vendorsaathi/internal/modules/auth/service"
	cartsmodel "vendorsaathi/internal/modules/carts/model"
	cartsrepository "vendorsaathi/internal/modules/carts/repository"
	cartsservice "vendorsaathi/internal/modules/carts/service"
	customermodel "vendorsaathi/internal/modules/customers/model"
	customersrepository "vendorsaathi/internal/modules/customers/repository"
	customersservice "vendorsaathi/internal/modules/customers/service"
	inventorymodel "vendorsaathi/internal/modules/inventory/model"
	inventoryrepository "vendorsaathi/internal/modules/inventory/repository"
	inventoryservice "vendorsaathi/internal/modules/inventory/service"
	notificationsrepository "vendorsaathi/internal/modules/notifications/repository"
	notificationsservice "vendorsaathi/internal/modules/notifications/service"
	ordersmodel "vendorsaathi/internal/modules/orders/model"
	ordersrepository "vendorsaathi/internal/modules/orders/repository"
	ordersservice "vendorsaathi/internal/modules/orders/service"
	paymentsmodel "vendorsaathi/internal/modules/payments/model"
	paymentsrepository "vendorsaathi/internal/modules/payments/repository"
	paymentsservice "vendorsaathi/internal/modules/payments/service"
	productsmodel "vendorsaathi/internal/modules/products/model"
	productsrepository "vendorsaathi/internal/modules/products/repository"
	productsservice "vendorsaathi/internal/modules/products/service"
	udhaarmodel "vendorsaathi/internal/modules/udhaar/model"
	udhaarrepository "vendorsaathi/internal/modules/udhaar/repository"
	udhaarservice "vendorsaathi/internal/modules/udhaar/service"
	vendorsmodel "vendorsaathi/internal/modules/vendors/model"
	vendorsrepository "vendorsaathi/internal/modules/vendors/repository"
	vendorsservice "vendorsaathi/internal/modules/vendors/service"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/platform/migration"
	"vendorsaathi/internal/shared"
)

func getTestDatabaseURL() string {
	url := os.Getenv("TEST_DATABASE_URL")
	if url != "" {
		return url
	}
	url = os.Getenv("DATABASE_URL")
	if url != "" {
		return url
	}
	return "postgres://postgres:postgres@localhost:5433/vendorsaathi_test?sslmode=disable"
}

func setupTestDB(t *testing.T) (*database.Pool, *database.TxManager) {
	t.Helper()
	dbURL := getTestDatabaseURL()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(dbURL))
	if err != nil {
		t.Skipf("skipping integration tests: unable to connect to database at %s: %v", dbURL, err)
		return nil, nil
	}

	// Apply migrations
	runner := migration.NewRunner(pool, "file://../migrations")
	migCtx, migCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migCancel()
	if err := runner.Up(migCtx); err != nil {
		// Try relative to current directory if running from repo root
		runnerRoot := migration.NewRunner(pool, "file://migrations")
		if errRoot := runnerRoot.Up(migCtx); errRoot != nil {
			t.Logf("migrations runner note: %v", errRoot)
		}
	}

	return pool, database.NewTxManager(pool)
}

func randomPhone() string {
	return fmt.Sprintf("+91%010d", rand.Int63n(8999999999)+1000000000)
}

func TestPostgresIntegration_ComprehensiveWorkflows(t *testing.T) {
	pool, txManager := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	// 1. Setup platform components
	jwtManager := auth.NewJWTManager("integration-test-secret-key-at-least-32-chars", "vendorsaathi-test", "vendorsaathi-clients", 15*time.Minute, 24*time.Hour)
	clock := database.RealClock{}

	authUserRepo := authrepository.NewUserRepository(pool)
	customerRepo := customersrepository.NewCustomerRepository(pool)
	authSvc := authservice.NewAuthService(authUserRepo, customerRepo, txManager, jwtManager, clock, pool, auth.PasswordParams{
		Memory:      1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})

	customerSvc := customersservice.NewCustomerService(customerRepo)
	vendorRepo := vendorsrepository.NewVendorRepository(pool)
	vendorSvc := vendorsservice.NewVendorService(vendorRepo)
	productRepo := productsrepository.NewProductRepository(pool)
	productSvc := productsservice.NewProductService(productRepo)
	inventoryRepo := inventoryrepository.NewInventoryRepository(pool)
	inventorySvc := inventoryservice.NewInventoryService(inventoryRepo)
	cartRepo := cartsrepository.NewCartRepository(pool)
	cartSvc := cartsservice.NewCartService(cartRepo, productRepo, inventoryRepo)
	orderRepo := ordersrepository.NewOrderRepository(pool)
	orderSvc := ordersservice.NewOrderService(pool, orderRepo, cartRepo, inventoryRepo, productRepo)
	paymentRepo := paymentsrepository.NewPaymentRepository(pool)
	paymentSvc := paymentsservice.NewPaymentService(paymentRepo)
	udhaarRepo := udhaarrepository.NewUdhaarRepository(pool)
	udhaarSvc := udhaarservice.NewUdhaarService(udhaarRepo)
	notifRepo := notificationsrepository.NewNotificationRepository(pool)
	notifSvc := notificationsservice.NewNotificationService(notifRepo)

	// ==========================================
	// 1. AUTHENTICATION & SECURITY WORKFLOW
	// ==========================================
	t.Run("Auth: Registration, Login, Token Rotation and Account Lockout", func(t *testing.T) {
		phone := randomPhone()
		password := "TestPass@12345"

		// Registration
		regTokens, err := authSvc.Register(ctx, authmodel.RegisterInput{
			Phone:    phone,
			Password: password,
		})
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
		if regTokens.AccessToken == "" || regTokens.RefreshToken == "" {
			t.Fatalf("Expected access and refresh tokens, got empty")
		}

		// Valid Login
		loginTokens, err := authSvc.Login(ctx, authmodel.LoginInput{
			Phone:    phone,
			Password: password,
		})
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if loginTokens.User.Phone != phone {
			t.Fatalf("Expected phone %s, got %s", phone, loginTokens.User.Phone)
		}

		// Validate Access Token Claims
		claims, err := jwtManager.ValidateAccessToken(loginTokens.AccessToken)
		if err != nil {
			t.Fatalf("AccessToken validation failed: %v", err)
		}
		if claims.UserID != loginTokens.User.ID.String() {
			t.Fatalf("Claims user ID mismatch: %s vs %s", claims.UserID, loginTokens.User.ID.String())
		}

		// Refresh Token Rotation
		refreshTokens, err := authSvc.Refresh(ctx, authmodel.RefreshInput{
			RefreshToken: loginTokens.RefreshToken,
		})
		if err != nil {
			t.Fatalf("Refresh token rotation failed: %v", err)
		}
		if refreshTokens.RefreshToken == loginTokens.RefreshToken {
			t.Fatalf("Refresh token was not rotated to a new token")
		}

		// Replay of old refresh token must fail (revoked)
		_, err = authSvc.Refresh(ctx, authmodel.RefreshInput{
			RefreshToken: loginTokens.RefreshToken,
		})
		if err == nil {
			t.Fatalf("Expected old refresh token replay to fail")
		}

		// Logout revokes the session
		err = authSvc.Logout(ctx, refreshTokens.RefreshToken)
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}

		// Refresh after logout must fail
		_, err = authSvc.Refresh(ctx, authmodel.RefreshInput{
			RefreshToken: refreshTokens.RefreshToken,
		})
		if err == nil {
			t.Fatalf("Expected refresh after logout to fail")
		}

		// Account Lockout after 5 failed login attempts
		lockPhone := randomPhone()
		_, _ = authSvc.Register(ctx, authmodel.RegisterInput{
			Phone:    lockPhone,
			Password: password,
		})

		for i := 1; i <= 5; i++ {
			_, err = authSvc.Login(ctx, authmodel.LoginInput{
				Phone:    lockPhone,
				Password: "WrongPassword!",
			})
			if err == nil {
				t.Fatalf("Expected failure on attempt %d", i)
			}
		}

		// 6th attempt must be locked out
		_, err = authSvc.Login(ctx, authmodel.LoginInput{
			Phone:    lockPhone,
			Password: password, // Even correct password must be rejected while locked
		})
		if err == nil {
			t.Fatalf("Expected account lockout after 5 failed attempts")
		}
		if appErr, ok := err.(*shared.AppError); !ok || appErr.Code != "ACCOUNT_LOCKED" {
			t.Logf("Got error code: %v", err)
		}
	})

	// ==========================================
	// 2. INVENTORY & TRANSACTION INTEGRITY
	// ==========================================
	t.Run("Inventory: Stock Adjustment, Movements, and Atomic Reservation", func(t *testing.T) {
		// Create Vendor User
		vendorPhone := randomPhone()
		vTokens, err := authSvc.Register(ctx, authmodel.RegisterInput{
			Phone:    vendorPhone,
			Password: "VendorPass@12345",
		})
		if err != nil {
			t.Fatalf("Vendor registration failed: %v", err)
		}
		vendorUserID := vTokens.User.ID

		// Grant VENDOR role
		vendorRole, err := authUserRepo.GetRoleByCode(ctx, "VENDOR")
		if err == nil && vendorRole != nil {
			_ = authUserRepo.AssignRole(ctx, vendorUserID, vendorRole.ID)
		}

		// Create Vendor Profile
		vProfile, err := vendorSvc.CreateProfile(ctx, vendorUserID, vendorsmodel.CreateVendorProfileInput{
			BusinessName: "Ramesh Organic Veggies",
			Phone:        vendorPhone,
		})
		if err != nil {
			t.Fatalf("Create vendor profile failed: %v", err)
		}

		// Create Product
		price := int64(12000) // ₹120.00
		cost := int64(8000)   // ₹80.00
		prod, err := productSvc.CreateProduct(ctx, vProfile.ID, productsmodel.CreateProductInput{
			Name:           "Fresh Organic Tomatoes",
			PricePaise:     price,
			CostPricePaise: &cost,
		})
		if err != nil {
			t.Fatalf("Create product failed: %v", err)
		}

		// Stock Adjustment: +100 units
		adjKey := uuid.New().String()
		inv, _, err := inventorySvc.AdjustStock(ctx, vProfile.ID, inventorymodel.AdjustStockInput{
			ProductID:      prod.ID,
			QtyDelta:       100,
			IdempotencyKey: &adjKey,
		})
		if err != nil {
			t.Fatalf("AdjustStock failed: %v", err)
		}
		if inv.OnHandQty != 100 || inv.AvailableQty() != 100 {
			t.Fatalf("Expected OnHandQty 100, got %d (available: %d)", inv.OnHandQty, inv.AvailableQty())
		}

		// Verify Movement Record
		movementsList, err := inventorySvc.GetMovements(ctx, prod.ID, vendorUserID, 20, "")
		if err != nil || len(movementsList.Data) == 0 {
			t.Fatalf("Expected inventory movements, got error: %v, count: %d", err, len(movementsList.Data))
		}
		if movementsList.Data[0].QtyDelta != 100 || movementsList.Data[0].MovementType != inventorymodel.MovementTypeAdjustment {
			t.Fatalf("Unexpected movement record: %+v", movementsList.Data[0])
		}

		// Insufficient Stock Error on oversell
		_, _, err = inventorySvc.AdjustStock(ctx, vProfile.ID, inventorymodel.AdjustStockInput{
			ProductID: prod.ID,
			QtyDelta:  -150, // More than on-hand
		})
		if err == nil {
			t.Fatalf("Expected error when reducing more than available stock")
		}

		// Concurrency Test: 5 concurrent reservations
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runErr := txManager.RunInTx(ctx, func(txCtx context.Context) error {
					refType := "CONCURRENT_ORDER"
					_, _, err := inventoryRepo.ReserveStock(txCtx, prod.ID, 10, nil, &refType, nil, nil)
					return err
				})
				if runErr == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		if successCount != 5 {
			t.Fatalf("Expected all 5 concurrent reservations to succeed with retry manager, succeeded: %d", successCount)
		}

		// Verify reserved stock
		latestInv, err := inventoryRepo.GetByProductID(ctx, prod.ID)
		if err != nil {
			t.Fatalf("GetByProductID failed: %v", err)
		}
		if latestInv.ReservedQty != 50 || latestInv.AvailableQty() != 50 {
			t.Fatalf("Expected reserved 50 and available 50, got reserved %d, available %d", latestInv.ReservedQty, latestInv.AvailableQty())
		}
	})

	// ==========================================
	// 3. CART, CHECKOUT, ORDER LIFECYCLE & PAYMENTS
	// ==========================================
	t.Run("Cart, Checkout, Order State Machine, Payments and Refunds", func(t *testing.T) {
		// Vendor setup
		vPhone := randomPhone()
		vTokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: vPhone, Password: "Pass@12345Vendor"})
		vUser := vTokens.User.ID
		vendorRole, _ := authUserRepo.GetRoleByCode(ctx, "VENDOR")
		if vendorRole != nil {
			_ = authUserRepo.AssignRole(ctx, vUser, vendorRole.ID)
		}
		vProfile, _ := vendorSvc.CreateProfile(ctx, vUser, vendorsmodel.CreateVendorProfileInput{
			BusinessName: "Bengaluru Daily Mart",
			Phone:        vPhone,
		})

		// Product setup: 20 units @ ₹50.00
		prod, err := productSvc.CreateProduct(ctx, vProfile.ID, productsmodel.CreateProductInput{
			Name:       "Fresh Spinach Bundle",
			PricePaise: 5000,
		})
		if err != nil {
			t.Fatalf("CreateProduct failed: %v", err)
		}
		_, _ = inventorySvc.AdjustStock(ctx, vProfile.ID, vUser, inventorymodel.AdjustStockInput{
			ProductID: prod.ID,
			QtyDelta:  20,
		})

		// Customer setup
		cPhone := randomPhone()
		cTokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: cPhone, Password: "Pass@12345Customer"})
		cUser := cTokens.User.ID
		_ = customerSvc.UpdateProfile(ctx, cUser, customermodel.UpdateCustomerProfileInput{
			FullName: "Anand Kumar",
		})

		// 3.1 Cart: Add item
		cart, err := cartSvc.AddItem(ctx, cUser, cartsmodel.AddItemInput{
			ProductID: prod.ID,
			Quantity:  2,
		})
		if err != nil {
			t.Fatalf("Cart AddItem failed: %v", err)
		}
		if len(cart.Items) != 1 || cart.Items[0].Quantity != 2 {
			t.Fatalf("Unexpected cart items: %+v", cart.Items)
		}

		// 3.2 Single-Vendor Constraint: attempt to add product from another vendor
		v2Phone := randomPhone()
		v2Tokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: v2Phone, Password: "Pass@12345Vendor2"})
		if vendorRole != nil {
			_ = authUserRepo.AssignRole(ctx, v2Tokens.User.ID, vendorRole.ID)
		}
		v2Profile, _ := vendorSvc.CreateProfile(ctx, v2Tokens.User.ID, vendorsmodel.CreateVendorProfileInput{
			BusinessName: "Mysore Sweet Corner",
			Phone:        v2Phone,
		})
		prod2, _ := productSvc.CreateProduct(ctx, v2Profile.ID, productsmodel.CreateProductInput{
			Name:       "Mysore Pak",
			PricePaise: 15000,
		})
		_, _ = inventorySvc.AdjustStock(ctx, v2Profile.ID, v2Tokens.User.ID, inventorymodel.AdjustStockInput{
			ProductID: prod2.ID,
			QtyDelta:  10,
		})

		_, err = cartSvc.AddItem(ctx, cUser, cartsmodel.AddItemInput{
			ProductID: prod2.ID,
			Quantity:  1,
		})
		if err == nil {
			t.Fatalf("Expected one-vendor cart constraint error when adding from second vendor")
		}

		// 3.3 Checkout with Idempotency
		checkoutKey := uuid.New().String()
		pickupAddr := "Storefront Pickup Counter"
		orderOut, err := orderSvc.Checkout(ctx, cUser, ordersmodel.CheckoutInput{
			FulfilmentType:        ordersmodel.FulfilmentTypePickup,
			PickupAddressSnapshot: &pickupAddr,
			IdempotencyKey:        checkoutKey,
		})
		if err != nil {
			t.Fatalf("Checkout failed: %v", err)
		}
		if orderOut.TotalPaise != 10000 { // 2 * ₹50.00
			t.Fatalf("Expected total 10000 paise, got %d", orderOut.TotalPaise)
		}
		if orderOut.Status != ordersmodel.OrderStatusPending {
			t.Fatalf("Expected initial status PENDING, got %s", orderOut.Status)
		}

		// Idempotent checkout returns the same order
		orderDup, err := orderSvc.Checkout(ctx, cUser, ordersmodel.CheckoutInput{
			FulfilmentType:        ordersmodel.FulfilmentTypePickup,
			PickupAddressSnapshot: &pickupAddr,
			IdempotencyKey:        checkoutKey,
		})
		if err != nil || orderDup.ID != orderOut.ID {
			t.Fatalf("Duplicate checkout with same key failed or returned different order ID")
		}

		// 3.4 Order State Machine: PENDING -> ACCEPTED -> PREPARING -> READY_FOR_PICKUP -> COMPLETED
		_, err = orderSvc.UpdateStatus(ctx, orderOut.ID, vUser, "VENDOR", ordersmodel.UpdateOrderStatusInput{
			Status: ordersmodel.OrderStatusAccepted,
		})
		if err != nil {
			t.Fatalf("Transition PENDING -> ACCEPTED failed: %v", err)
		}

		_, err = orderSvc.UpdateStatus(ctx, orderOut.ID, vUser, "VENDOR", ordersmodel.UpdateOrderStatusInput{
			Status: ordersmodel.OrderStatusPreparing,
		})
		if err != nil {
			t.Fatalf("Transition ACCEPTED -> PREPARING failed: %v", err)
		}

		_, err = orderSvc.UpdateStatus(ctx, orderOut.ID, vUser, "VENDOR", ordersmodel.UpdateOrderStatusInput{
			Status: ordersmodel.OrderStatusReadyForPickup,
		})
		if err != nil {
			t.Fatalf("Transition PREPARING -> READY_FOR_PICKUP failed: %v", err)
		}

		// Invalid transition: READY_FOR_PICKUP -> CANCELLED should fail
		_, err = orderSvc.UpdateStatus(ctx, orderOut.ID, vUser, "VENDOR", ordersmodel.UpdateOrderStatusInput{
			Status: ordersmodel.OrderStatusCancelled,
		})
		if err == nil {
			t.Fatalf("Invalid transition READY_FOR_PICKUP -> CANCELLED succeeded, expected error")
		}

		// Complete order
		finalOrder, err := orderSvc.UpdateStatus(ctx, orderOut.ID, vUser, "VENDOR", ordersmodel.UpdateOrderStatusInput{
			Status: ordersmodel.OrderStatusCompleted,
		})
		if err != nil || finalOrder.Status != ordersmodel.OrderStatusCompleted {
			t.Fatalf("Transition to COMPLETED failed: %v", err)
		}

		// 3.5 Payments & Overpayment Guard
		payKey := uuid.New().String()
		pm, err := paymentSvc.RecordPayment(ctx, vUser, paymentsmodel.RecordPaymentInput{
			OrderID:        orderOut.ID,
			AmountPaise:    6000, // ₹60 partial
			Method:         paymentsmodel.PaymentMethodCash,
			IdempotencyKey: &payKey,
		})
		if err != nil {
			t.Fatalf("RecordPayment failed: %v", err)
		}
		if pm.AmountPaise != 6000 {
			t.Fatalf("Expected payment 6000, got %d", pm.AmountPaise)
		}

		// Overpayment prevention: order total is 10000, already paid 6000, attempting 5000 (total 11000)
		_, err = paymentSvc.RecordPayment(ctx, vUser, paymentsmodel.RecordPaymentInput{
			OrderID:     orderOut.ID,
			AmountPaise: 5000,
			Method:      paymentsmodel.PaymentMethodCash,
		})
		if err == nil {
			t.Fatalf("Expected overpayment rejection, but payment succeeded")
		}

		// Pay exact balance (4000)
		_, err = paymentSvc.RecordPayment(ctx, vUser, paymentsmodel.RecordPaymentInput{
			OrderID:     orderOut.ID,
			AmountPaise: 4000,
			Method:      paymentsmodel.PaymentMethodUPI,
		})
		if err != nil {
			t.Fatalf("Paying remaining balance failed: %v", err)
		}

		// 3.6 Refund & Refund Limits
		refKey := uuid.New().String()
		refReason := "Customer returned 1 damaged bundle"
		refund, err := paymentSvc.CreateRefund(ctx, vUser, paymentsmodel.CreateRefundInput{
			OrderID:        orderOut.ID,
			AmountPaise:    5000,
			Reason:         &refReason,
			IdempotencyKey: &refKey,
		})
		if err != nil {
			t.Fatalf("CreateRefund failed: %v", err)
		}
		if refund.AmountPaise != 5000 {
			t.Fatalf("Expected refund amount 5000, got %d", refund.AmountPaise)
		}

		// Excess refund rejection: total paid is 10000, refunded 5000, remaining refundable is 5000
		_, err = paymentSvc.CreateRefund(ctx, vUser, paymentsmodel.CreateRefundInput{
			OrderID:     orderOut.ID,
			AmountPaise: 6000, // Exceeds 5000 refundable
		})
		if err == nil {
			t.Fatalf("Expected excess refund rejection, but refund succeeded")
		}
	})

	// ==========================================
	// 4. UDHAAR (CREDIT LEDGER) INTEGRITY
	// ==========================================
	t.Run("Udhaar: Account Creation, Credit, Repayment and Limit Integrity", func(t *testing.T) {
		vPhone := randomPhone()
		vTokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: vPhone, Password: "Pass@12345Vendor"})
		vUser := vTokens.User.ID
		vendorRole, _ := authUserRepo.GetRoleByCode(ctx, "VENDOR")
		if vendorRole != nil {
			_ = authUserRepo.AssignRole(ctx, vUser, vendorRole.ID)
		}
		vProfile, _ := vendorSvc.CreateProfile(ctx, vUser, vendorsmodel.CreateVendorProfileInput{
			BusinessName: "Vijayanagar Provision Store",
			Phone:        vPhone,
		})

		cPhone := randomPhone()
		cTokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: cPhone, Password: "Pass@12345Customer"})
		cUser := cTokens.User.ID

		// Create Udhaar Account
		creditLimit := int64(100000) // ₹1,000 credit limit
		uAcc, err := udhaarSvc.CreateAccount(ctx, vUser, udhaarmodel.CreateUdhaarAccountInput{
			CustomerID:       cUser,
			CreditLimitPaise: &creditLimit,
		})
		if err != nil {
			t.Fatalf("CreateAccount failed: %v", err)
		}
		if uAcc.BalancePaise != 0 {
			t.Fatalf("Expected initial balance 0, got %d", uAcc.BalancePaise)
		}

		// Record CREDIT entry (+ ₹350.00 = 35000 paise)
		credDesc := "Grocery items on credit"
		entry1, err := udhaarSvc.AddEntry(ctx, vUser, uAcc.ID, udhaarmodel.AddUdhaarEntryInput{
			EntryType:   udhaarmodel.UdhaarEntryTypeCredit,
			AmountPaise: 35000,
			Description: &credDesc,
		})
		if err != nil {
			t.Fatalf("AddEntry CREDIT failed: %v", err)
		}
		if entry1.AmountPaise != 35000 {
			t.Fatalf("Expected entry amount 35000, got %d", entry1.AmountPaise)
		}

		// Verify Account Balance
		accUpdated, err := udhaarSvc.GetAccount(ctx, vUser, uAcc.ID)
		if err != nil || accUpdated.BalancePaise != 35000 {
			t.Fatalf("Expected updated balance 35000, got %d (err: %v)", accUpdated.BalancePaise, err)
		}

		// Repayment limit check: attempting to repay 40000 when debt is 35000
		_, err = udhaarSvc.AddEntry(ctx, vUser, uAcc.ID, udhaarmodel.AddUdhaarEntryInput{
			EntryType:   udhaarmodel.UdhaarEntryTypeRepayment,
			AmountPaise: 40000,
		})
		if err == nil {
			t.Fatalf("Expected repayment limit exceeded error, but repayment succeeded")
		}

		// Valid Repayment of 20000 paise (₹200.00)
		repayDesc := "Partial cash payment"
		_, err = udhaarSvc.AddEntry(ctx, vUser, uAcc.ID, udhaarmodel.AddUdhaarEntryInput{
			EntryType:   udhaarmodel.UdhaarEntryTypeRepayment,
			AmountPaise: 20000,
			Description: &repayDesc,
		})
		if err != nil {
			t.Fatalf("AddEntry REPAYMENT failed: %v", err)
		}

		// Verify Remaining Balance (15000 paise)
		accFinal, err := udhaarSvc.GetAccount(ctx, cUser, uAcc.ID) // Also tests customer access
		if err != nil || accFinal.BalancePaise != 15000 {
			t.Fatalf("Expected final balance 15000, got %d (err: %v)", accFinal.BalancePaise, err)
		}

		// Ledger Append-Only Integrity: entries list must contain exactly 2 entries in chronological order
		entries, _, _, err := udhaarSvc.GetEntries(ctx, vUser, uAcc.ID, 20, "")
		if err != nil {
			t.Fatalf("GetEntries failed: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("Expected 2 ledger entries, got %d", len(entries))
		}
	})

	// ==========================================
	// 5. NOTIFICATIONS & SYNC WORKFLOW
	// ==========================================
	t.Run("Notifications: Creation, Unread Count, and Device Token Registration", func(t *testing.T) {
		phone := randomPhone()
		tokens, _ := authSvc.Register(ctx, authmodel.RegisterInput{Phone: phone, Password: "Pass@12345Notif"})
		user := tokens.User.ID

		// Send Notification
		err := notifSvc.NotifyUser(ctx, user, "ORDER_UPDATE", "Order Accepted", "Your order has been accepted by Ramesh Fruits", map[string]interface{}{
			"order_id": uuid.New().String(),
		})
		if err != nil {
			t.Fatalf("NotifyUser failed: %v", err)
		}

		// Fetch notifications
		list, err := notifSvc.GetMyNotifications(ctx, user, 10, "")
		if err != nil {
			t.Fatalf("GetMyNotifications failed: %v", err)
		}
		if len(list.Data) != 1 || list.UnreadCount != 1 {
			t.Fatalf("Expected 1 unread notification, got count %d, unread %d", len(list.Data), list.UnreadCount)
		}

		// Mark read
		notifID := list.Data[0].ID
		err = notifSvc.MarkRead(ctx, notifID, user)
		if err != nil {
			t.Fatalf("MarkRead failed: %v", err)
		}

		// Verify unread count is 0
		listAfter, err := notifSvc.GetMyNotifications(ctx, user, 10, "")
		if err != nil || listAfter.UnreadCount != 0 {
			t.Fatalf("Expected unread count 0 after mark read, got %d", listAfter.UnreadCount)
		}

		// Register Device Token
		deviceToken, err := notifSvc.RegisterDeviceToken(ctx, user, "fcm-test-token-device-xyz", "ANDROID")
		if err != nil {
			t.Fatalf("RegisterDeviceToken failed: %v", err)
		}
		if deviceToken.Token != "fcm-test-token-device-xyz" {
			t.Fatalf("Expected token 'fcm-test-token-device-xyz', got %s", deviceToken.Token)
		}
	})
}
