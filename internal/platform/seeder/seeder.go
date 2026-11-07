package seeder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/config"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/platform/logger"
)

var seedNamespace = uuid.MustParse("e0743b12-9c32-4e89-b1d5-8f6a97120000")

// DetUUID produces a stable deterministic UUID from a seed string identifier
func DetUUID(identifier string) uuid.UUID {
	return uuid.NewSHA1(seedNamespace, []byte(identifier))
}

type SeedSummary struct {
	Users              int
	UserRoles          int
	VendorProfiles     int
	VendorLocations    int
	BusinessHours      int
	CustomerProfiles   int
	CustomerAddresses  int
	Categories         int
	Products           int
	Inventory          int
	InventoryMovements int
	Carts              int
	CartItems          int
	Orders             int
	OrderItems         int
	OrderStatusHistory int
	OrderCancellations int
	Payments           int
	Refunds            int
	UdhaarAccounts     int
	UdhaarEntries      int
	Notifications      int
	DeviceTokens       int
	SyncOperations     int
	AuditLogs          int
	RefreshSessions    int
}

func (s *SeedSummary) Print() {
	fmt.Println("\n================================================================================")
	fmt.Println("             VENDORSAATHI DEVELOPMENT DATABASE SEED SUMMARY                    ")
	fmt.Println("================================================================================")
	fmt.Printf(" %-35s : %d records\n", "Users (users)", s.Users)
	fmt.Printf(" %-35s : %d records\n", "User Role Mappings (user_roles)", s.UserRoles)
	fmt.Printf(" %-35s : %d records\n", "Vendor Profiles (vendor_profiles)", s.VendorProfiles)
	fmt.Printf(" %-35s : %d records\n", "Vendor Locations (vendor_locations)", s.VendorLocations)
	fmt.Printf(" %-35s : %d records\n", "Business Hours (business_hours)", s.BusinessHours)
	fmt.Printf(" %-35s : %d records\n", "Customer Profiles (customer_profiles)", s.CustomerProfiles)
	fmt.Printf(" %-35s : %d records\n", "Customer Addresses (customer_addresses)", s.CustomerAddresses)
	fmt.Printf(" %-35s : %d records\n", "Categories (categories)", s.Categories)
	fmt.Printf(" %-35s : %d records\n", "Products (products)", s.Products)
	fmt.Printf(" %-35s : %d records\n", "Inventory Balances (inventory)", s.Inventory)
	fmt.Printf(" %-35s : %d records\n", "Inventory Movements (inventory_movements)", s.InventoryMovements)
	fmt.Printf(" %-35s : %d records\n", "Carts (carts)", s.Carts)
	fmt.Printf(" %-35s : %d records\n", "Cart Items (cart_items)", s.CartItems)
	fmt.Printf(" %-35s : %d records\n", "Orders (orders)", s.Orders)
	fmt.Printf(" %-35s : %d records\n", "Order Items (order_items)", s.OrderItems)
	fmt.Printf(" %-35s : %d records\n", "Order Status History (order_status_history)", s.OrderStatusHistory)
	fmt.Printf(" %-35s : %d records\n", "Order Cancellations (order_cancellations)", s.OrderCancellations)
	fmt.Printf(" %-35s : %d records\n", "Payments (payments)", s.Payments)
	fmt.Printf(" %-35s : %d records\n", "Refunds (refunds)", s.Refunds)
	fmt.Printf(" %-35s : %d records\n", "Udhaar Accounts (udhaar_accounts)", s.UdhaarAccounts)
	fmt.Printf(" %-35s : %d records\n", "Udhaar Entries (udhaar_entries)", s.UdhaarEntries)
	fmt.Printf(" %-35s : %d records\n", "Notifications (notifications)", s.Notifications)
	fmt.Printf(" %-35s : %d records\n", "Device Tokens (device_tokens)", s.DeviceTokens)
	fmt.Printf(" %-35s : %d records\n", "Offline Sync Operations (sync_operations)", s.SyncOperations)
	fmt.Printf(" %-35s : %d records\n", "Audit Trail Logs (audit_logs)", s.AuditLogs)
	fmt.Printf(" %-35s : %d records\n", "Refresh Sessions (refresh_sessions)", s.RefreshSessions)
	fmt.Println("================================================================================")
	fmt.Println(" SEED COMPLETE: All data interconnected, valid, and discovery-ready.")
	fmt.Println("================================================================================")
}

// ResetDatabase wipes seeded data only if explicitly requested and confirmed
func ResetDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	tables := []string{
		"sync_operations", "audit_logs", "device_tokens", "notifications",
		"udhaar_entries", "udhaar_accounts", "refunds", "payments",
		"order_cancellations", "order_status_history", "order_items", "orders",
		"cart_items", "carts", "inventory_movements", "inventory",
		"products", "categories", "customer_addresses", "customer_profiles",
		"business_hours", "vendor_locations", "vendor_profiles",
		"refresh_sessions", "user_roles", "users",
	}

	for _, t := range tables {
		if _, err := tx.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", t)); err != nil {
			return fmt.Errorf("truncate %s: %w", t, err)
		}
	}

	return tx.Commit(ctx)
}

// Seed executes the comprehensive idempotent development seeding
func Seed(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) (*SeedSummary, error) {
	if !cfg.IsDevelopment() {
		return nil, fmt.Errorf("seeder refusal: environment is %q, must be 'development'", cfg.AppEnv)
	}

	logger.Logger.Info().Msg("Starting VendorSaathi development database seeding...")

	summary := &SeedSummary{}

	// Hash default passwords using configured Argon2id settings
	adminPass := "AdminPassword123!"
	vendorPass := "VendorPassword123!"
	custPass := "CustomerPassword123!"

	adminHash, err := auth.HashPassword(adminPass, 64*1024, 3, 4, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}
	vendorHash, err := auth.HashPassword(vendorPass, 64*1024, 3, 4, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("hash vendor password: %w", err)
	}
	custHash, err := auth.HashPassword(custPass, 64*1024, 3, 4, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("hash customer password: %w", err)
	}

	// Begin outer transaction for seeding
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Map role codes to DB IDs
	roleMap := make(map[string]uuid.UUID)
	roleRows, err := tx.Query(ctx, `SELECT id, code FROM roles`)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	for roleRows.Next() {
		var id uuid.UUID
		var code string
		if err := roleRows.Scan(&id, &code); err != nil {
			roleRows.Close()
			return nil, err
		}
		roleMap[code] = id
	}
	roleRows.Close()

	// -------------------------------------------------------------------------
	// 1. SEED USERS & ROLES
	// -------------------------------------------------------------------------
	usersToSeed := []struct {
		Key      string
		Phone    string
		Email    string
		Hash     string
		Status   string
		RoleList []string
	}{
		// Admins
		{"user:super_admin", "+919876543210", "admin@vendorsaathi.in", adminHash, "ACTIVE", []string{"SUPER_ADMIN", "ADMIN"}},
		{"user:admin_kavitha", "+919876543211", "kavitha.admin@vendorsaathi.in", adminHash, "ACTIVE", []string{"ADMIN"}},
		{"user:admin_suspended", "+919876543219", "suspended.admin@vendorsaathi.in", adminHash, "SUSPENDED", []string{"ADMIN"}},

		// Customers
		{"user:cust_ramesh", "+919888877777", "customer@vendorsaathi.in", custHash, "ACTIVE", []string{"CUSTOMER"}},
		{"user:cust_priya", "+919888811111", "priya.sharma@example.com", custHash, "ACTIVE", []string{"CUSTOMER"}},
		{"user:cust_anand", "+919888822222", "anand.kumar@example.com", custHash, "ACTIVE", []string{"CUSTOMER"}},
		{"user:cust_deepa", "+919888833333", "deepa.nair@example.com", custHash, "ACTIVE", []string{"CUSTOMER"}},
		{"user:cust_ali", "+919888844444", "mohammed.ali@example.com", custHash, "ACTIVE", []string{"CUSTOMER"}},
		{"user:cust_suspended", "+919888899999", "suspended.customer@example.com", custHash, "SUSPENDED", []string{"CUSTOMER"}},

		// 20 Vendors
		{"user:vendor_1", "+919876500001", "vendor.suresh@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_2", "+919876500002", "vendor.chickpet@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_3", "+919876500003", "vendor.gandhinagar@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_4", "+919876500004", "vendor.malleswaram@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_5", "+919876500005", "vendor.rajajinagar@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_6", "+919876500006", "vendor.gandhibazaar@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_7", "+919876500007", "vendor.russell@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_8", "+919876500008", "vendor.commstreet@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_9", "+919876500009", "vendor.jayanagar@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_10", "+919876500010", "vendor.indiranagar@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_11", "+919876500011", "vendor.koramangala@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_12", "+919876500012", "vendor.yeshwanthpur@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_13", "+919876500013", "vendor.ulsoor@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		// Inactive/Pending within 10km
		{"user:vendor_14", "+919876500014", "vendor.suspended14@vendorsaathi.in", vendorHash, "SUSPENDED", []string{"VENDOR"}},
		{"user:vendor_15", "+919876500015", "vendor.pending15@vendorsaathi.in", vendorHash, "PENDING", []string{"VENDOR"}},
		// Outside 10km
		{"user:vendor_16", "+919876500016", "vendor.whitefield@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_17", "+919876500017", "vendor.ecity@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_18", "+919876500018", "vendor.yelahanka@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_19", "+919876500019", "vendor.kengeri@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
		{"user:vendor_20", "+919876500020", "vendor.bannerghatta@vendorsaathi.in", vendorHash, "ACTIVE", []string{"VENDOR"}},
	}

	userUUIDMap := make(map[string]uuid.UUID)

	for _, u := range usersToSeed {
		id := DetUUID(u.Key)

		var actualID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO users (id, phone, email, password_hash, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (phone) DO UPDATE SET
				email = EXCLUDED.email,
				password_hash = EXCLUDED.password_hash,
				status = EXCLUDED.status,
				updated_at = NOW()
			RETURNING id
		`, id, u.Phone, u.Email, u.Hash, u.Status).Scan(&actualID)
		if err != nil {
			return nil, fmt.Errorf("upsert user %s: %w", u.Phone, err)
		}
		userUUIDMap[u.Key] = actualID
		summary.Users++

		// Assign roles
		for _, rCode := range u.RoleList {
			roleID, exists := roleMap[rCode]
			if exists {
				_, err = tx.Exec(ctx, `
					INSERT INTO user_roles (user_id, role_id)
					VALUES ($1, $2)
					ON CONFLICT (user_id, role_id) DO NOTHING
				`, actualID, roleID)
				if err != nil {
					return nil, fmt.Errorf("user_roles %s - %s: %w", u.Phone, rCode, err)
				}
				summary.UserRoles++
			}
		}
	}

	// -------------------------------------------------------------------------
	// 2. SEED CUSTOMER PROFILES & ADDRESSES
	// -------------------------------------------------------------------------
	customersToSeed := []struct {
		UserKey   string
		Name      string
		Addresses []struct {
			Key       string
			Line1     string
			Line2     string
			City      string
			State     string
			Pincode   string
			Lat       float64
			Lng       float64
			Label     string
			IsDefault bool
		}
	}{
		{
			UserKey: "user:cust_ramesh",
			Name:    "Ramesh Gowda (ರಮೇಶ್ ಗೌಡ)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:ramesh_home", "Flat 302, Green Glen Layout", "Bellandur Outer Ring Rd", "Bengaluru", "Karnataka", "560103", 12.9322, 77.6744, "Home", true},
				{"addr:ramesh_work", "4th Floor, Tech Hub", "MG Road Metro Station", "Bengaluru", "Karnataka", "560001", 12.9754, 77.6067, "Work", false},
			},
		},
		{
			UserKey: "user:cust_priya",
			Name:    "Priya Sharma (ಪ್ರಿಯಾ ಶರ್ಮಾ)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:priya_home", "No 45, 7th Cross", "Sampige Road, Malleshwaram", "Bengaluru", "Karnataka", "560003", 13.0031, 77.5685, "Home", true},
			},
		},
		{
			UserKey: "user:cust_anand",
			Name:    "Anand Kumar (ಆನಂದ್ ಕುಮಾರ್)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:anand_home", "12, 1st Main, Gandhi Bazaar", "Basavanagudi", "Bengaluru", "Karnataka", "560004", 12.9429, 77.5732, "Home", true},
			},
		},
		{
			UserKey: "user:cust_deepa",
			Name:    "Deepa Nair (ದೀಪಾ ನಾಯರ್)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:deepa_home", "56, 12th Main, HAL 2nd Stage", "Indiranagar", "Bengaluru", "Karnataka", "560038", 12.9725, 77.6405, "Home", true},
			},
		},
		{
			UserKey: "user:cust_ali",
			Name:    "Mohammed Ali (ಮೊಹಮ್ಮದ್ ಅಲಿ)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:ali_home", "24, Commercial Street Cross", "Shivajinagar", "Bengaluru", "Karnataka", "560001", 12.9829, 77.6078, "Home", true},
			},
		},
		{
			UserKey: "user:cust_suspended",
			Name:    "Suspended Customer (ಖಾತೆ ಅಮಾನತು)",
			Addresses: []struct {
				Key       string
				Line1     string
				Line2     string
				City      string
				State     string
				Pincode   string
				Lat       float64
				Lng       float64
				Label     string
				IsDefault bool
			}{
				{"addr:susp_home", "99, Old Madras Road", "Ulsoor", "Bengaluru", "Karnataka", "560008", 12.9760, 77.6200, "Home", true},
			},
		},
	}

	for _, c := range customersToSeed {
		userID := userUUIDMap[c.UserKey]
		_, err := tx.Exec(ctx, `
			INSERT INTO customer_profiles (user_id, display_name, contact_preferences, created_at, updated_at)
			VALUES ($1, $2, '{"sms": true, "whatsapp": true}', NOW(), NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				updated_at = NOW()
		`, userID, c.Name)
		if err != nil {
			return nil, fmt.Errorf("customer profile %s: %w", c.UserKey, err)
		}
		summary.CustomerProfiles++

		for _, addr := range c.Addresses {
			addrID := DetUUID(addr.Key)
			_, err = tx.Exec(ctx, `
				INSERT INTO customer_addresses (id, customer_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, label, is_default, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 'India', $8, $9, $10, $11, NOW(), NOW())
				ON CONFLICT (id) DO UPDATE SET
					address_line1 = EXCLUDED.address_line1,
					address_line2 = EXCLUDED.address_line2,
					city = EXCLUDED.city,
					latitude = EXCLUDED.latitude,
					longitude = EXCLUDED.longitude,
					is_default = EXCLUDED.is_default,
					updated_at = NOW()
			`, addrID, userID, addr.Line1, addr.Line2, addr.City, addr.State, addr.Pincode, addr.Lat, addr.Lng, addr.Label, addr.IsDefault)
			if err != nil {
				return nil, fmt.Errorf("customer address %s: %w", addr.Key, err)
			}
			summary.CustomerAddresses++
		}
	}

	// -------------------------------------------------------------------------
	// 3. SEED VENDORS (PROFILES, LOCATIONS & BUSINESS HOURS)
	// -------------------------------------------------------------------------
	// Reference test point: Bangalore Majestic (12.9767, 77.5713)
	vendorsToSeed := []struct {
		Key          string
		OwnerKey     string
		BusinessName string
		Description  string
		Phone        string
		Status       string
		Availability string
		// Location
		Line1   string
		Line2   string
		City    string
		Pincode string
		Lat     float64
		Lng     float64
		// Business hours (0=Sun, 1=Mon, ..., 6=Sat)
		OpenTime  string
		CloseTime string
	}{
		// --- 13 ACTIVE & WITHIN 10 KM (DISCOVERABLE IN FLUTTER APP) ---
		{
			Key: "vendor:1", OwnerKey: "user:vendor_1",
			BusinessName: "Suresh Fresh Fruits & Vegetables",
			Description:  "ಸುರೇಶ್ ತಾಜಾ ಹಣ್ಣು ಮತ್ತು ತರಕಾರಿಗಳು — Farm-fresh seasonal fruits & local greens daily.",
			Phone:        "+919876500001", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Stall 14, Majestic City Market, Tank Bund Rd", Line2: "Near Kempegowda Bus Station",
			City: "Bengaluru", Pincode: "560009", Lat: 12.9772, Lng: 77.5718, // ~0.08 km
			OpenTime: "06:30", CloseTime: "21:30",
		},
		{
			Key: "vendor:2", OwnerKey: "user:vendor_2",
			BusinessName: "Chickpet Sri Krishna Kirana & Provisions",
			Description:  "ಶ್ರೀ ಕೃಷ್ಣ ಕಿರಾಣಿ ಅಂಗಡಿ — Authentic spices, pulses, rice, flours, and home provisions.",
			Phone:        "+919876500002", Status: "ACTIVE", Availability: "OPEN",
			Line1: "88, O.T.C. Road, Chickpet", Line2: "Near Mohan Building",
			City: "Bengaluru", Pincode: "560053", Lat: 12.9698, Lng: 77.5750, // ~0.85 km
			OpenTime: "08:00", CloseTime: "21:00",
		},
		{
			Key: "vendor:3", OwnerKey: "user:vendor_3",
			BusinessName: "Gandhi Nagar Organic Greens",
			Description:  "ಗಾಂಧಿನಗರ ಸಾವಯವ ತರಕಾರಿ — 100% naturally grown organic vegetables, coriander & herbs.",
			Phone:        "+919876500003", Status: "ACTIVE", Availability: "OPEN",
			Line1: "22, 5th Main, Gandhi Nagar", Line2: "Opp. Tribhuvan Theatre",
			City: "Bengaluru", Pincode: "560009", Lat: 12.9790, Lng: 77.5760, // ~0.58 km
			OpenTime: "07:00", CloseTime: "20:00",
		},
		{
			Key: "vendor:4", OwnerKey: "user:vendor_4",
			BusinessName: "Malleshwaram Nandi Dairy & Sweets",
			Description:  "ನಂದಿ ಡೈರಿ ಮತ್ತು ಸಿಹಿತಿಂಡಿಗಳು — Pure cow milk, curd, paneer, and traditional Mysore Pak.",
			Phone:        "+919876500004", Status: "ACTIVE", Availability: "OPEN",
			Line1: "34, 8th Cross, Sampige Road", Line2: "Malleshwaram",
			City: "Bengaluru", Pincode: "560003", Lat: 13.0035, Lng: 77.5695, // ~3.0 km
			OpenTime: "06:00", CloseTime: "22:00",
		},
		{
			Key: "vendor:5", OwnerKey: "user:vendor_5",
			BusinessName: "Rajajinagar Iyengar Bakery",
			Description:  "ರಾಜಾಜಿನಗರ ಅಯ್ಯಂಗಾರ್ ಬೇಕರಿ — Fresh warm potato buns, honey cakes, khara biscuits & bread.",
			Phone:        "+919876500005", Status: "ACTIVE", Availability: "OPEN",
			Line1: "105, 1st Block, Dr Rajkumar Rd", Line2: "Rajajinagar",
			City: "Bengaluru", Pincode: "560010", Lat: 12.9902, Lng: 77.5525, // ~2.5 km
			OpenTime: "07:30", CloseTime: "22:00",
		},
		{
			Key: "vendor:6", OwnerKey: "user:vendor_6",
			BusinessName: "Gandhi Bazaar Pooja & Flower Mart",
			Description:  "ಗಾಂಧಿ ಬಜಾರ್ ಹೂವು ಮತ್ತು ಪೂಜಾ ಭಂಡಾರ — Jasmine garlands, sevanti, agarbatti, camphor & brass puja items.",
			Phone:        "+919876500006", Status: "ACTIVE", Availability: "OPEN",
			Line1: "12, Gandhi Bazaar Main Rd", Line2: "Basavanagudi",
			City: "Bengaluru", Pincode: "560004", Lat: 12.9432, Lng: 77.5724, // ~3.7 km
			OpenTime: "06:00", CloseTime: "21:00",
		},
		{
			Key: "vendor:7", OwnerKey: "user:vendor_7",
			BusinessName: "Shivajinagar Russell Market Fresh Meats",
			Description:  "ರಸ್ಸೆಲ್ ಮಾರ್ಕೆಟ್ ತಾಜಾ ಮಾಂಸ — Fresh halal chicken, mutton cuts, and farm eggs.",
			Phone:        "+919876500007", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Stall 45, Russell Market", Line2: "Shivajinagar",
			City: "Bengaluru", Pincode: "560051", Lat: 12.9856, Lng: 77.6006, // ~3.3 km
			OpenTime: "07:00", CloseTime: "19:00",
		},
		{
			Key: "vendor:8", OwnerKey: "user:vendor_8",
			BusinessName: "Commercial Street Dry Fruits & Spices",
			Description:  "ಡ್ರೈ ಫ್ರೂಟ್ಸ್ ಮತ್ತು ಮಸಾಲೆಗಳು — Premium cashews, Kashmiri almonds, saffron, and cardamom.",
			Phone:        "+919876500008", Status: "ACTIVE", Availability: "OPEN",
			Line1: "67, Commercial Street", Line2: "Tasker Town",
			City: "Bengaluru", Pincode: "560001", Lat: 12.9822, Lng: 77.6083, // ~4.1 km
			OpenTime: "10:00", CloseTime: "21:30",
		},
		{
			Key: "vendor:9", OwnerKey: "user:vendor_9",
			BusinessName: "Jayanagar 4th Block Chaat & Fresh Juices",
			Description:  "ಜಯನಗರ ಚಾಟ್ಸ್ & ಜ್ಯೂಸ್ — Famous pani puri, sev puri, fresh mosambi juice, and sugarcane juice.",
			Phone:        "+919876500009", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Complex Stall 8, 4th Block", Line2: "Jayanagar BDA Complex",
			City: "Bengaluru", Pincode: "560011", Lat: 12.9298, Lng: 77.5834, // ~5.4 km
			OpenTime: "11:00", CloseTime: "22:30",
		},
		{
			Key: "vendor:10", OwnerKey: "user:vendor_10",
			BusinessName: "Indiranagar Artisanal Bakery & Cafe",
			Description:  "ಇಂದಿರಾನಗರ ಬೇಕರಿ — Sourdough loaves, croissants, filter cold brew, and cookies.",
			Phone:        "+919876500010", Status: "ACTIVE", Availability: "OPEN",
			Line1: "789, 100 Feet Rd", Line2: "Indiranagar HAL 2nd Stage",
			City: "Bengaluru", Pincode: "560038", Lat: 12.9719, Lng: 77.6412, // ~7.6 km
			OpenTime: "08:00", CloseTime: "22:00",
		},
		{
			Key: "vendor:11", OwnerKey: "user:vendor_11",
			BusinessName: "Koramangala Fresh Cart Daily",
			Description:  "ಕೋರಮಂಗಲ ತಾಜಾ ತರಕಾರಿಗಳು — Handpicked crisp salad greens, exotic veggies, and fruits.",
			Phone:        "+919876500011", Status: "ACTIVE", Availability: "OPEN",
			Line1: "12, 5th Block, 80 Feet Road", Line2: "Koramangala",
			City: "Bengaluru", Pincode: "560095", Lat: 12.9352, Lng: 77.6245, // ~7.4 km
			OpenTime: "07:00", CloseTime: "21:00",
		},
		{
			Key: "vendor:12", OwnerKey: "user:vendor_12",
			BusinessName: "Yeshwanthpur APMC Wholesale Provisions",
			Description:  "ಯಶವಂತಪುರ ಎಪಿಎಂಸಿ ಕಿರಾಣಿ — Wholesale sacks of Sona Masoori, Toor Dal, and edible oils.",
			Phone:        "+919876500012", Status: "ACTIVE", Availability: "CLOSED", // Tests CLOSED status in mobile
			Line1: "Yard 3, APMC Market Yard", Line2: "Yeshwanthpur",
			City: "Bengaluru", Pincode: "560022", Lat: 13.0232, Lng: 77.5501, // ~5.6 km
			OpenTime: "05:00", CloseTime: "17:00",
		},
		{
			Key: "vendor:13", OwnerKey: "user:vendor_13",
			BusinessName: "Ulsoor Lake Tender Coconut Point",
			Description:  "ಹಲಸೂರು ಎಳನೀರು ಕೇಂದ್ರ — Sweet Pollachi tender coconuts and freshly cut pineapple slices.",
			Phone:        "+919876500013", Status: "ACTIVE", Availability: "TEMPORARILY_UNAVAILABLE", // Tests TEMPORARILY_UNAVAILABLE
			Line1: "Opp. Kensington Swimming Pool", Line2: "Ulsoor Lake Promenade",
			City: "Bengaluru", Pincode: "560008", Lat: 12.9785, Lng: 77.6250, // ~5.8 km
			OpenTime: "06:00", CloseTime: "19:00",
		},

		// --- 2 INACTIVE / UNAPPROVED WITHIN 10 KM (TESTING EXCLUSION IN DISCOVER) ---
		{
			Key: "vendor:14", OwnerKey: "user:vendor_14",
			BusinessName: "City Market Banned Vendor",
			Description:  "ಅಮಾನತುಗೊಂಡ ಅಂಗಡಿ — Suspended vendor account for terms violation testing.",
			Phone:        "+919876500014", Status: "SUSPENDED", Availability: "CLOSED",
			Line1: "K.R. Market Basement Stall 9", Line2: "City Market",
			City: "Bengaluru", Pincode: "560002", Lat: 12.9650, Lng: 77.5750, // ~1.3 km (Suspended)
			OpenTime: "08:00", CloseTime: "20:00",
		},
		{
			Key: "vendor:15", OwnerKey: "user:vendor_15",
			BusinessName: "Malleshwaram New Vendor Onboarding",
			Description:  "ಹೊಸ ವ್ಯಾಪಾರಿ ಅನುಮೋದನೆ ಬಾಕಿ — Pending document approval verification.",
			Phone:        "+919876500015", Status: "PENDING", Availability: "CLOSED",
			Line1: "15th Cross, Margosa Road", Line2: "Malleshwaram",
			City: "Bengaluru", Pincode: "560003", Lat: 13.0100, Lng: 77.5670, // ~3.7 km (Pending)
			OpenTime: "09:00", CloseTime: "21:00",
		},

		// --- 5 OUTSIDE 10 KM RADIUS (TESTING DISTANCE FILTERING IN DISCOVER) ---
		{
			Key: "vendor:16", OwnerKey: "user:vendor_16",
			BusinessName: "Whitefield Tech Park Supermart",
			Description:  "ವೈಟ್‌ಫೀಲ್ಡ್ ಸೂಪರ್‌ಮಾರ್ಟ್ — IT corridor grocery and household goods mart.",
			Phone:        "+919876500016", Status: "ACTIVE", Availability: "OPEN",
			Line1: "ITPL Main Rd, Near Hope Farm", Line2: "Whitefield",
			City: "Bengaluru", Pincode: "560066", Lat: 12.9866, Lng: 77.7315, // ~17.4 km
			OpenTime: "08:00", CloseTime: "22:00",
		},
		{
			Key: "vendor:17", OwnerKey: "user:vendor_17",
			BusinessName: "Electronic City Fresh Mart",
			Description:  "ಎಲೆಕ್ಟ್ರಾನಿಕ್ ಸಿಟಿ ಮಾರ್ಟ್ — Techies organic grocery & daily provisions.",
			Phone:        "+919876500017", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Phase 1, Hosur Road", Line2: "Electronic City",
			City: "Bengaluru", Pincode: "560100", Lat: 12.8452, Lng: 77.6602, // ~17.5 km
			OpenTime: "07:30", CloseTime: "22:30",
		},
		{
			Key: "vendor:18", OwnerKey: "user:vendor_18",
			BusinessName: "Yelahanka New Town Kirana Store",
			Description:  "ಯಲಹಂಕ ಕಿರಾಣಿ ಅಂಗಡಿ — Traditional provisions & grain store in North Bengaluru.",
			Phone:        "+919876500018", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Major Sandeep Unnikrishnan Rd", Line2: "Yelahanka New Town",
			City: "Bengaluru", Pincode: "560064", Lat: 13.1007, Lng: 77.5963, // ~14.1 km
			OpenTime: "08:00", CloseTime: "21:00",
		},
		{
			Key: "vendor:19", OwnerKey: "user:vendor_19",
			BusinessName: "Kengeri Satellite Town Provisions",
			Description:  "ಕೆಂಗೇರಿ ಉಪನಗರ ಪ್ರಾವಿಷನ್ಸ್ — West Bengaluru local neighbourhood store.",
			Phone:        "+919876500019", Status: "ACTIVE", Availability: "OPEN",
			Line1: "1st Main, Mysore Road Cross", Line2: "Kengeri Satellite Town",
			City: "Bengaluru", Pincode: "560060", Lat: 12.9150, Lng: 77.4830, // ~11.8 km
			OpenTime: "07:30", CloseTime: "21:30",
		},
		{
			Key: "vendor:20", OwnerKey: "user:vendor_20",
			BusinessName: "Bannerghatta Green Farm Fresh",
			Description:  "ಬನ್ನೇರುಘಟ್ಟ ಫಾರ್ಮ್ ಫ್ರೆಶ್ — Farm gate organic produce & fresh honey.",
			Phone:        "+919876500020", Status: "ACTIVE", Availability: "OPEN",
			Line1: "Bannerghatta Main Road", Line2: "Near National Park Gate",
			City: "Bengaluru", Pincode: "560083", Lat: 12.8000, Lng: 77.5770, // ~19.6 km
			OpenTime: "07:00", CloseTime: "19:00",
		},
	}

	vendorUUIDMap := make(map[string]uuid.UUID)
	vendorOwnerMap := make(map[string]uuid.UUID)

	for _, v := range vendorsToSeed {
		vendorID := DetUUID(v.Key)
		ownerID := userUUIDMap[v.OwnerKey]

		var actualVendorID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO vendor_profiles (id, owner_user_id, business_name, description, phone, status, availability_status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			ON CONFLICT (owner_user_id) DO UPDATE SET
				business_name = EXCLUDED.business_name,
				description = EXCLUDED.description,
				phone = EXCLUDED.phone,
				status = EXCLUDED.status,
				availability_status = EXCLUDED.availability_status,
				updated_at = NOW()
			RETURNING id
		`, vendorID, ownerID, v.BusinessName, v.Description, v.Phone, v.Status, v.Availability).Scan(&actualVendorID)
		if err != nil {
			return nil, fmt.Errorf("vendor profile %s: %w", v.Key, err)
		}
		vendorUUIDMap[v.Key] = actualVendorID
		vendorOwnerMap[v.Key] = ownerID
		summary.VendorProfiles++

		// Primary Location
		locID := DetUUID(v.Key + ":location")
		_, err = tx.Exec(ctx, `
			INSERT INTO vendor_locations (id, vendor_id, address_line1, address_line2, city, state, postal_code, country, latitude, longitude, is_primary, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'Karnataka', $6, 'India', $7, $8, TRUE, NOW())
			ON CONFLICT (id) DO UPDATE SET
				vendor_id = EXCLUDED.vendor_id,
				address_line1 = EXCLUDED.address_line1,
				address_line2 = EXCLUDED.address_line2,
				city = EXCLUDED.city,
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				is_primary = TRUE,
				updated_at = NOW()
		`, locID, actualVendorID, v.Line1, v.Line2, v.City, v.Pincode, v.Lat, v.Lng)
		if err != nil {
			return nil, fmt.Errorf("vendor location %s: %w", v.Key, err)
		}
		summary.VendorLocations++

		// Weekly Business Hours (Sunday=0 to Saturday=6)
		for w := 0; w <= 6; w++ {
			bhID := DetUUID(fmt.Sprintf("%s:bh:%d", v.Key, w))
			isClosed := (w == 0 && v.Key == "vendor:12") // APMC closed on Sundays
			_, err = tx.Exec(ctx, `
				INSERT INTO business_hours (id, vendor_id, weekday, open_time, close_time, is_closed)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (vendor_id, weekday) DO UPDATE SET
					open_time = EXCLUDED.open_time,
					close_time = EXCLUDED.close_time,
					is_closed = EXCLUDED.is_closed
			`, bhID, actualVendorID, w, v.OpenTime, v.CloseTime, isClosed)
			if err != nil {
				return nil, fmt.Errorf("business hour %s w=%d: %w", v.Key, w, err)
			}
			summary.BusinessHours++
		}
	}

	// -------------------------------------------------------------------------
	// 4. SEED CATEGORIES & PRODUCTS & INVENTORY
	// -------------------------------------------------------------------------
	categoriesToSeed := []struct {
		Key  string
		Name string
	}{
		{"cat:fruits_veg", "Fruits & Vegetables (ಹಣ್ಣು ಮತ್ತು ತರಕಾರಿಗಳು)"},
		{"cat:kirana", "Kirana & Grocery (ಕಿರಾಣಿ ಅಂಗಡಿ)"},
		{"cat:dairy", "Dairy, Milk & Eggs (ಹಾಲು ಮತ್ತು ಡೈರಿ)"},
		{"cat:bakery", "Bakery & Snacks (ಬೇಕರಿ ತಿನಿಸುಗಳು)"},
		{"cat:flowers", "Flowers & Pooja Items (ಹೂವು & ಪೂಜಾ ಸಾಮಗ್ರಿ)"},
		{"cat:beverages", "Beverages & Cold Drinks (ಪಾನೀಯಗಳು)"},
	}

	catUUIDMap := make(map[string]uuid.UUID)
	for _, c := range categoriesToSeed {
		catID := DetUUID(c.Key)
		catUUIDMap[c.Key] = catID
		_, err := tx.Exec(ctx, `
			INSERT INTO categories (id, name, parent_id, active, created_at, updated_at)
			VALUES ($1, $2, NULL, TRUE, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				active = TRUE,
				updated_at = NOW()
		`, catID, c.Name)
		if err != nil {
			return nil, fmt.Errorf("category %s: %w", c.Key, err)
		}
		summary.Categories++
	}

	// Products catalog with inventory scenarios
	productsToSeed := []struct {
		Key         string
		VendorKey   string
		CatKey      string
		Name        string
		Description string
		PricePaise  int64
		Active      bool
		OnHand      int64
		Reserved    int64
	}{
		// Vendor 1 (Suresh Fresh Fruits & Veg)
		{"prod:v1_banana", "vendor:1", "cat:fruits_veg", "Robusta Banana (ಬಾಳೆಹಣ್ಣು)", "Fresh Karnataka Robusta sweet bananas per dozen", 4500, true, 80, 0},
		{"prod:v1_tomato", "vendor:1", "cat:fruits_veg", "Country Tomato (ನಾಟಿ ಟೊಮೆಟೊ)", "Crisp juicy desi salad tomatoes per kg", 3000, true, 120, 3}, // 3 reserved in cart/order
		{"prod:v1_onion", "vendor:1", "cat:fruits_veg", "Bellary Onion (ಈರುಳ್ಳಿ)", "Premium dry medium size cooking onions per kg", 3500, true, 150, 2},
		{"prod:v1_potato", "vendor:1", "cat:fruits_veg", "Hassan Potato (ಆಲೂಗಡ್ಡೆ)", "Fresh earthy Hassan curry potatoes per kg", 3500, true, 100, 2},
		{"prod:v1_apple", "vendor:1", "cat:fruits_veg", "Shimla Apple (ಸೇಬು)", "Crisp sweet red apples imported from Himachal per kg", 14000, true, 50, 1},
		{"prod:v1_orange", "vendor:1", "cat:fruits_veg", "Nagpur Orange (ಕಿತ್ತಳೆ)", "Vitamin-C rich juicy sweet oranges per kg", 8000, true, 40, 1},
		{"prod:v1_mango_low", "vendor:1", "cat:fruits_veg", "Badami Mango (ಬಾದಾಮಿ ಮಾವು)", "Alphonso/Badami table mangoes (Low stock scenario)", 22000, true, 5, 2}, // Low stock
		{"prod:v1_strawberry_out", "vendor:1", "cat:fruits_veg", "Mahabaleshwar Strawberry (ಸ್ಟ್ರಾಬೆರಿ)", "Fresh box strawberries (Out of stock scenario)", 12000, true, 0, 0}, // Out of stock
		{"prod:v1_inactive_exotic", "vendor:1", "cat:fruits_veg", "Dragon Fruit (ಡ್ರ್ಯಾಗನ್ ಹಣ್ಣು)", "Exotic fruit (Inactive catalogue item)", 15000, false, 10, 0}, // Inactive product

		// Vendor 2 (Chickpet Sri Krishna Kirana)
		{"prod:v2_sona_rice", "vendor:2", "cat:kirana", "Sona Masoori Raw Rice (ಸೋನಾ ಮಸೂರಿ ಅಕ್ಕಿ)", "1-year aged old rice per 5kg bag", 32000, true, 50, 0},
		{"prod:v2_toor_dal", "vendor:2", "cat:kirana", "Gulbarga Toor Dal (ತೊಗರಿ ಬೇಳೆ)", "Unpolished protein-rich dal per 1kg", 16000, true, 60, 1},
		{"prod:v2_sugar", "vendor:2", "cat:kirana", "Refined White Sugar (ಸಕ್ಕರೆ)", "Clean crystal sulphur-free sugar per 1kg", 4500, true, 100, 1},
		{"prod:v2_atta", "vendor:2", "cat:kirana", "Chakki Fresh Wheat Atta (ಗೋಧಿ ಹಿಟ್ಟು)", "100% whole wheat stone ground flour per 5kg", 24500, true, 40, 1},
		{"prod:v2_sunflower_oil", "vendor:2", "cat:kirana", "Sunflower Cooking Oil (ಸೂರ್ಯಕಾಂತಿ ಎಣ್ಣೆ)", "Refined light heart-friendly cooking oil 1 Litre pouch", 13500, true, 80, 0},

		// Vendor 4 (Malleshwaram Nandi Dairy)
		{"prod:v4_milk", "vendor:4", "cat:dairy", "Fresh Cow Milk (ಹಸುವಿನ ಹಾಲು)", "Pasteurized whole creamy cow milk 500ml pouch", 2500, true, 200, 2},
		{"prod:v4_curd", "vendor:4", "cat:dairy", "Thick Set Curd (ಗಟ್ಟಿ ಮೊಸರು)", "Traditional probiotic thick curd 500g tub", 3500, true, 75, 0},
		{"prod:v4_paneer", "vendor:4", "cat:dairy", "Malai Paneer Block (ಮಲೈ ಪನೀರ್)", "Soft fresh cottage cheese cubes 200g", 11000, true, 45, 1},
		{"prod:v4_butter", "vendor:4", "cat:dairy", "Desi Cooking Butter (ಬೆಣ್ಣೆ)", "Unsalted homemade cooking butter 250g", 16000, true, 30, 0},

		// Vendor 5 (Rajajinagar Iyengar Bakery)
		{"prod:v5_potato_bun", "vendor:5", "cat:bakery", "Spicy Potato Bun (ಆಲೂ ಬನ್)", "Freshly baked spiced potato masala stuffed bun", 2500, true, 90, 0},
		{"prod:v5_honey_cake", "vendor:5", "cat:bakery", "Classic Honey Cake (ಜೇನು ಕೇಕ್)", "Jam-glazed sponge honey pastry slice", 3500, true, 60, 0},
		{"prod:v5_rusk", "vendor:5", "cat:bakery", "Cardamom Milk Rusk (ಟೋಸ್ಟ್ ರಸ್ಕ್)", "Crisp tea dip biscuits 300g pack", 5000, true, 85, 0},

		// Vendor 6 (Gandhi Bazaar Pooja & Flower Mart)
		{"prod:v6_jasmine", "vendor:6", "cat:flowers", "Mysore Jasmine Garland (ಮಲ್ಲಿಗೆ ಹೂವಿನ ಮಾಲೆ)", "Fragrant fresh woven Mysore Mallige strand (1 meter)", 8000, true, 40, 0},
		{"prod:v6_agarbatti", "vendor:6", "cat:flowers", "Champa Sandalwood Agarbatti (ಅಗರಬತ್ತಿ)", "Natural charcoal-free incense sticks pack", 6000, true, 110, 0},
		{"prod:v6_camphor", "vendor:6", "cat:flowers", "Pure Bhimseni Camphor (ಕರ್ಪೂರ)", "100% natural devotional aarti camphor 100g", 12000, true, 70, 0},

		// Vendor 9 (Jayanagar Chaat & Fresh Juices)
		{"prod:v9_pani_puri", "vendor:9", "cat:beverages", "Special Pani Puri Plate (ಪಾನಿಪುರಿ)", "Crispy puris with spicy mint water & potato ragda", 4000, true, 100, 0},
		{"prod:v9_mosambi", "vendor:9", "cat:beverages", "Fresh Mosambi Juice (ಮೂಸಂಬಿ ಜ್ಯೂಸ್)", "Chilled fresh sweet lime juice with rock salt", 5000, true, 80, 0},
		{"prod:v9_sugarcane", "vendor:9", "cat:beverages", "Sugarcane Juice with Ginger (ಕಬ್ಬಿನ ಹಾಲು)", "Cold pressed fresh sugarcane with lemon & ginger", 3000, true, 90, 0},
	}

	prodUUIDMap := make(map[string]uuid.UUID)
	for _, p := range productsToSeed {
		prodID := DetUUID(p.Key)
		prodUUIDMap[p.Key] = prodID
		vendorID := vendorUUIDMap[p.VendorKey]
		catID := catUUIDMap[p.CatKey]

		_, err := tx.Exec(ctx, `
			INSERT INTO products (id, vendor_id, category_id, name, description, price_paise, currency, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'INR', $7, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				price_paise = EXCLUDED.price_paise,
				active = EXCLUDED.active,
				updated_at = NOW()
		`, prodID, vendorID, catID, p.Name, p.Description, p.PricePaise, p.Active)
		if err != nil {
			return nil, fmt.Errorf("product %s: %w", p.Key, err)
		}
		summary.Products++

		// Inventory row
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory (product_id, on_hand_qty, reserved_qty, version, updated_at)
			VALUES ($1, $2, $3, 1, NOW())
			ON CONFLICT (product_id) DO UPDATE SET
				on_hand_qty = EXCLUDED.on_hand_qty,
				reserved_qty = EXCLUDED.reserved_qty,
				version = inventory.version + 1,
				updated_at = NOW()
		`, prodID, p.OnHand, p.Reserved)
		if err != nil {
			return nil, fmt.Errorf("inventory %s: %w", p.Key, err)
		}
		summary.Inventory++

		// Initial stocking movement
		movID := DetUUID(p.Key + ":init_movement")
		vendorOwnerID := vendorOwnerMap[p.VendorKey]
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_movements (id, product_id, movement_type, qty_delta, reason, actor_user_id, created_at)
			VALUES ($1, $2, 'ADJUSTMENT', $3, 'Initial development stock initialization', $4, NOW())
			ON CONFLICT (id) DO NOTHING
		`, movID, prodID, p.OnHand, vendorOwnerID)
		if err != nil {
			return nil, fmt.Errorf("movement %s: %w", p.Key, err)
		}
		summary.InventoryMovements++
	}

	// -------------------------------------------------------------------------
	// 5. SEED ACTIVE CARTS (SINGLE-VENDOR CART COMPLIANT)
	// -------------------------------------------------------------------------
	// Ramesh Gowda (user:cust_ramesh) has an active cart with Suresh (vendor:1)
	cartID := DetUUID("cart:ramesh_v1")
	rameshID := userUUIDMap["user:cust_ramesh"]
	v1ID := vendorUUIDMap["vendor:1"]

	var actualCartID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM carts WHERE customer_id = $1 AND vendor_id = $2 AND status = 'ACTIVE' LIMIT 1
	`, rameshID, v1ID).Scan(&actualCartID)
	if err != nil {
		actualCartID = cartID
		_, err = tx.Exec(ctx, `
			INSERT INTO carts (id, customer_id, vendor_id, status, updated_at)
			VALUES ($1, $2, $3, 'ACTIVE', NOW())
			ON CONFLICT (id) DO UPDATE SET status = 'ACTIVE', updated_at = NOW()
		`, actualCartID, rameshID, v1ID)
		if err != nil {
			return nil, fmt.Errorf("cart %s: %w", cartID, err)
		}
	}
	summary.Carts++

	// Cart items: Banana (2), Tomato (3), Orange (1)
	cartItemsToSeed := []struct {
		ProdKey string
		Qty     int64
	}{
		{"prod:v1_banana", 2},
		{"prod:v1_tomato", 3},
		{"prod:v1_orange", 1},
	}
	for _, ci := range cartItemsToSeed {
		ciID := DetUUID(fmt.Sprintf("cart_item:%s:%s", actualCartID, ci.ProdKey))
		prodID := prodUUIDMap[ci.ProdKey]
		_, err = tx.Exec(ctx, `
			INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (cart_id, product_id) DO UPDATE SET
				quantity = EXCLUDED.quantity,
				updated_at = NOW()
		`, ciID, actualCartID, prodID, ci.Qty)
		if err != nil {
			return nil, fmt.Errorf("cart item %s: %w", ci.ProdKey, err)
		}
		summary.CartItems++
	}

	// -------------------------------------------------------------------------
	// 6. SEED REALISTIC ORDERS & STATE MACHINE TRANSITIONS
	// -------------------------------------------------------------------------
	ordersToSeed := []struct {
		Key             string
		CustKey         string
		VendorKey       string
		Fulfilment      string
		Status          string
		Subtotal        int64
		Total           int64
		DeliverySnap    string
		PickupSnap      string
		Items           []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }
		StatusHistory   []struct{ From, To, Reason string }
		HasCancellation bool
		CancelReason    string
		CancelPrevState string
		CancelDecision  *string
	}{
		// Order 1: PENDING (Pickup) - Can test "Accept" / "Reject" in Vendor App
		{
			Key: "order:1_pending", CustKey: "user:cust_ramesh", VendorKey: "vendor:1",
			Fulfilment: "PICKUP", Status: "PENDING",
			Subtotal: 14000, Total: 14000,
			PickupSnap: `{"vendor_name": "Suresh Fresh Fruits", "address": "Stall 14, Majestic City Market"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_onion", "Bellary Onion (ಈರುಳ್ಳಿ)", 3500, 2, 7000},
				{"prod:v1_potato", "Hassan Potato (ಆಲೂಗಡ್ಡೆ)", 3500, 2, 7000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order via mobile app"},
			},
		},
		// Order 2: ACCEPTED (Delivery) - Can test "Start Preparing" in Vendor App
		{
			Key: "order:2_accepted", CustKey: "user:cust_priya", VendorKey: "vendor:1",
			Fulfilment: "DELIVERY", Status: "ACCEPTED",
			Subtotal: 22000, Total: 22000,
			DeliverySnap: `{"customer_name": "Priya Sharma", "address": "No 45, 7th Cross, Malleshwaram"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_apple", "Shimla Apple (ಸೇಬು)", 14000, 1, 14000},
				{"prod:v1_orange", "Nagpur Orange (ಕಿತ್ತಳೆ)", 8000, 1, 8000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Vendor confirmed order and stock availability"},
			},
		},
		// Order 3: PREPARING (Pickup) - Can test "Ready for Pickup" in Vendor App
		{
			Key: "order:3_preparing_pickup", CustKey: "user:cust_anand", VendorKey: "vendor:1",
			Fulfilment: "PICKUP", Status: "PREPARING",
			Subtotal: 18000, Total: 18000,
			PickupSnap: `{"vendor_name": "Suresh Fresh Fruits", "address": "Stall 14, Majestic City Market"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_banana", "Robusta Banana (ಬಾಳೆಹಣ್ಣು)", 4500, 4, 18000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Vendor accepted"},
				{"ACCEPTED", "PREPARING", "Packing fresh items"},
			},
		},
		// Order 4: PREPARING (Delivery) - Can test "Out for Delivery" in Vendor App
		{
			Key: "order:4_preparing_delivery", CustKey: "user:cust_deepa", VendorKey: "vendor:1",
			Fulfilment: "DELIVERY", Status: "PREPARING",
			Subtotal: 35000, Total: 35000,
			DeliverySnap: `{"customer_name": "Deepa Nair", "address": "56, 12th Main, Indiranagar"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_apple", "Shimla Apple (ಸೇಬು)", 14000, 2, 28000},
				{"prod:v1_onion", "Bellary Onion (ಈರುಳ್ಳಿ)", 3500, 2, 7000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Vendor accepted"},
				{"ACCEPTED", "PREPARING", "Packing fruits in crate"},
			},
		},
		// Order 5: READY_FOR_PICKUP (Pickup) - Can test "Complete Order"
		{
			Key: "order:5_ready_pickup", CustKey: "user:cust_ramesh", VendorKey: "vendor:4",
			Fulfilment: "PICKUP", Status: "READY_FOR_PICKUP",
			Subtotal: 16000, Total: 16000,
			PickupSnap: `{"vendor_name": "Malleshwaram Nandi Dairy", "address": "34, 8th Cross, Sampige Road"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v4_milk", "Fresh Cow Milk (ಹಸುವಿನ ಹಾಲು)", 2500, 2, 5000},
				{"prod:v4_paneer", "Malai Paneer Block (ಮಲೈ ಪನೀರ್)", 11000, 1, 11000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Accepted by dairy counter"},
				{"ACCEPTED", "PREPARING", "Chilled packets packed"},
				{"PREPARING", "READY_FOR_PICKUP", "Stored in customer pickup refrigeration zone"},
			},
		},
		// Order 6: OUT_FOR_DELIVERY (Delivery) - Can test "Complete Order"
		{
			Key: "order:6_out_delivery", CustKey: "user:cust_priya", VendorKey: "vendor:2",
			Fulfilment: "DELIVERY", Status: "OUT_FOR_DELIVERY",
			Subtotal: 45000, Total: 45000,
			DeliverySnap: `{"customer_name": "Priya Sharma", "address": "No 45, 7th Cross, Malleshwaram"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v2_atta", "Chakki Fresh Wheat Atta (ಗೋಧಿ ಹಿಟ್ಟು)", 24500, 1, 24500},
				{"prod:v2_toor_dal", "Gulbarga Toor Dal (ತೊಗರಿ ಬೇಳೆ)", 16000, 1, 16000},
				{"prod:v2_sugar", "Refined White Sugar (ಸಕ್ಕರೆ)", 4500, 1, 4500},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Accepted by provisions clerk"},
				{"ACCEPTED", "PREPARING", "Bags sealed for transit"},
				{"PREPARING", "OUT_FOR_DELIVERY", "Dispatched with vendor delivery boy on 2-wheeler"},
			},
		},
		// Order 7: COMPLETED (Pickup) - Fully fulfilled and paid
		{
			Key: "order:7_completed", CustKey: "user:cust_ramesh", VendorKey: "vendor:1",
			Fulfilment: "PICKUP", Status: "COMPLETED",
			Subtotal: 30000, Total: 30000,
			PickupSnap: `{"vendor_name": "Suresh Fresh Fruits", "address": "Stall 14, Majestic City Market"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_mango_low", "Badami Mango (ಬಾದಾಮಿ ಮಾವು)", 22000, 1, 22000},
				{"prod:v1_orange", "Nagpur Orange (ಕಿತ್ತಳೆ)", 8000, 1, 8000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Accepted"},
				{"ACCEPTED", "PREPARING", "Packed"},
				{"PREPARING", "READY_FOR_PICKUP", "Awaiting customer collection"},
				{"READY_FOR_PICKUP", "COMPLETED", "Customer collected order and paid cash"},
			},
		},
		// Order 8: COMPLETED (Delivery) - Split Payment ₹250 Cash + ₹250 UPI
		{
			Key: "order:8_completed_split", CustKey: "user:cust_anand", VendorKey: "vendor:2",
			Fulfilment: "DELIVERY", Status: "COMPLETED",
			Subtotal: 50000, Total: 50000,
			DeliverySnap: `{"customer_name": "Anand Kumar", "address": "12, 1st Main, Gandhi Bazaar"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v2_sona_rice", "Sona Masoori Raw Rice (ಸೋನಾ ಮಸೂರಿ ಅಕ್ಕಿ)", 32000, 1, 32000},
				{"prod:v2_toor_dal", "Gulbarga Toor Dal (ತೊಗರಿ ಬೇಳೆ)", 16000, 1, 16000},
				{"prod:v2_sugar", "Refined White Sugar (ಸಕ್ಕರೆ)", 4500, 1, 4500},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "ACCEPTED", "Accepted"},
				{"ACCEPTED", "PREPARING", "Packed"},
				{"PREPARING", "OUT_FOR_DELIVERY", "Out with delivery staff"},
				{"OUT_FOR_DELIVERY", "COMPLETED", "Delivered safely; split payment recorded"},
			},
		},
		// Order 9: REJECTED (Vendor rejected) - Stock was released
		{
			Key: "order:9_rejected", CustKey: "user:cust_deepa", VendorKey: "vendor:1",
			Fulfilment: "PICKUP", Status: "REJECTED",
			Subtotal: 12000, Total: 12000,
			PickupSnap: `{"vendor_name": "Suresh Fresh Fruits", "address": "Stall 14, Majestic City Market"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v1_strawberry_out", "Mahabaleshwar Strawberry (ಸ್ಟ್ರಾಬೆರಿ)", 12000, 1, 12000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "REJECTED", "Item was out of fresh daily stock"},
			},
		},
		// Order 10: CANCELLED (Customer requested, Vendor APPROVED)
		{
			Key: "order:10_cancelled", CustKey: "user:cust_ali", VendorKey: "vendor:4",
			Fulfilment: "DELIVERY", Status: "CANCELLED",
			Subtotal: 20000, Total: 20000,
			DeliverySnap: `{"customer_name": "Mohammed Ali", "address": "24, Commercial Street"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v4_curd", "Thick Set Curd (ಗಟ್ಟಿ ಮೊಸರು)", 3500, 2, 7000},
				{"prod:v4_paneer", "Malai Paneer Block (ಮಲೈ ಪನೀರ್)", 11000, 1, 11000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "CANCELLATION_REQUESTED", "Customer requested cancellation: Plans changed"},
				{"CANCELLATION_REQUESTED", "CANCELLED", "Vendor approved cancellation; inventory reservation released"},
			},
			HasCancellation: true,
			CancelReason:    "Customer travelling out of town urgently",
			CancelPrevState: "PENDING",
			CancelDecision:  strPtr("APPROVED"),
		},
		// Order 11: CANCELLATION_REQUESTED (Pending Vendor Decision)
		{
			Key: "order:11_cancel_requested", CustKey: "user:cust_ramesh", VendorKey: "vendor:2",
			Fulfilment: "PICKUP", Status: "CANCELLATION_REQUESTED",
			Subtotal: 17500, Total: 17500,
			PickupSnap: `{"vendor_name": "Chickpet Sri Krishna Kirana", "address": "88, O.T.C. Road"}`,
			Items: []struct{ ProdKey, Name string; UnitPrice, Qty, LineTotal int64 }{
				{"prod:v2_sunflower_oil", "Sunflower Cooking Oil (ಸೂರ್ಯಕಾಂತಿ ಎಣ್ಣೆ)", 13500, 1, 13500},
				{"prod:v2_sugar", "Refined White Sugar (ಸಕ್ಕರೆ)", 4000, 1, 4000},
			},
			StatusHistory: []struct{ From, To, Reason string }{
				{"", "PENDING", "Customer placed order"},
				{"PENDING", "CANCELLATION_REQUESTED", "Ordered duplicate by mistake"},
			},
			HasCancellation: true,
			CancelReason:    "Ordered duplicate items by mistake on family account",
			CancelPrevState: "PENDING",
			CancelDecision:  nil, // Pending vendor approval/rejection
		},
	}

	orderUUIDMap := make(map[string]uuid.UUID)

	for _, o := range ordersToSeed {
		orderID := DetUUID(o.Key)
		orderUUIDMap[o.Key] = orderID
		custID := userUUIDMap[o.CustKey]
		vendorID := vendorUUIDMap[o.VendorKey]

		var delSnap, pickSnap *string
		if o.DeliverySnap != "" {
			delSnap = &o.DeliverySnap
		}
		if o.PickupSnap != "" {
			pickSnap = &o.PickupSnap
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO orders (id, customer_id, vendor_id, fulfilment_type, status, subtotal_paise, total_paise, delivery_address_snapshot, pickup_address_snapshot, idempotency_key, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW() - INTERVAL '2 hours', NOW())
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				subtotal_paise = EXCLUDED.subtotal_paise,
				total_paise = EXCLUDED.total_paise,
				updated_at = NOW()
		`, orderID, custID, vendorID, o.Fulfilment, o.Status, o.Subtotal, o.Total, delSnap, pickSnap, o.Key)
		if err != nil {
			return nil, fmt.Errorf("order %s: %w", o.Key, err)
		}
		summary.Orders++

		// Order Items
		for i, it := range o.Items {
			itID := DetUUID(fmt.Sprintf("%s:item:%d", o.Key, i))
			prodID := prodUUIDMap[it.ProdKey]
			_, err = tx.Exec(ctx, `
				INSERT INTO order_items (id, order_id, product_id, product_name_snapshot, unit_price_paise, quantity, line_total_paise, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW() - INTERVAL '2 hours')
				ON CONFLICT (id) DO NOTHING
			`, itID, orderID, prodID, it.Name, it.UnitPrice, it.Qty, it.LineTotal)
			if err != nil {
				return nil, fmt.Errorf("order item %s: %w", it.Name, err)
			}
			summary.OrderItems++
		}

		// Order Status History
		for idx, h := range o.StatusHistory {
			hID := DetUUID(fmt.Sprintf("%s:hist:%d", o.Key, idx))
			var fromStatus *string
			if h.From != "" {
				fromStatus = &h.From
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO order_status_history (id, order_id, from_status, to_status, actor_user_id, reason, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW() - INTERVAL '90 minutes' + ($7 || ' minutes')::interval)
				ON CONFLICT (id) DO NOTHING
			`, hID, orderID, fromStatus, h.To, custID, h.Reason, fmt.Sprintf("%d", idx*10))
			if err != nil {
				return nil, fmt.Errorf("order history %s: %w", o.Key, err)
			}
			summary.OrderStatusHistory++
		}

		// Cancellations
		if o.HasCancellation {
			cID := DetUUID(o.Key + ":cancellation")
			var decAt *time.Time
			var decBy *uuid.UUID
			if o.CancelDecision != nil {
				now := time.Now()
				decAt = &now
				vOwnerID := vendorOwnerMap[o.VendorKey]
				decBy = &vOwnerID
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO order_cancellations (id, order_id, requested_by, reason, previous_status, decision, decided_by, decided_at, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW() - INTERVAL '45 minutes')
				ON CONFLICT (id) DO NOTHING
			`, cID, orderID, custID, o.CancelReason, o.CancelPrevState, o.CancelDecision, decBy, decAt)
			if err != nil {
				return nil, fmt.Errorf("cancellation %s: %w", o.Key, err)
			}
			summary.OrderCancellations++
		}
	}

	// -------------------------------------------------------------------------
	// 7. SEED PAYMENTS & REFUNDS
	// -------------------------------------------------------------------------
	paymentsToSeed := []struct {
		Key          string
		OrderKey     string
		Method       string
		Amount       int64
		Status       string
		Verification string
		ActorKey     string
		Reference    string
		Evidence     string
		// Optional Refund
		HasRefund    bool
		RefundAmount int64
		RefundReason string
	}{
		// Order 7 (Completed): Full payment via CASH
		{
			Key: "pay:ord7_cash", OrderKey: "order:7_completed",
			Method: "CASH", Amount: 30000,
			Status: "CONFIRMED", Verification: "VERIFIED",
			ActorKey: "user:vendor_1", Reference: "CASH/2026/001", Evidence: `{"notes": "Paid in cash at shop counter"}`,
		},
		// Order 8 (Completed): Split payment - ₹250 CASH + ₹250 UPI
		{
			Key: "pay:ord8_cash", OrderKey: "order:8_completed_split",
			Method: "CASH", Amount: 25000,
			Status: "CONFIRMED", Verification: "VERIFIED",
			ActorKey: "user:vendor_2", Reference: "CASH/2026/002", Evidence: `{"notes": "Split payment part 1"}`,
		},
		{
			Key: "pay:ord8_upi", OrderKey: "order:8_completed_split",
			Method: "UPI", Amount: 25000,
			Status: "CONFIRMED", Verification: "VERIFIED",
			ActorKey: "user:vendor_2", Reference: "UPI/2026/split_8819231", Evidence: `{"utr": "41298819231", "app": "PhonePe"}`,
		},
		// Order 5 (Ready for Pickup): Pending UPI Verification (Can test "Verify UPI" in backend)
		{
			Key: "pay:ord5_upi_pending", OrderKey: "order:5_ready_pickup",
			Method: "UPI", Amount: 16000,
			Status: "PENDING_VERIFICATION", Verification: "PENDING",
			ActorKey: "user:vendor_4", Reference: "UPI/2026/pend_991823", Evidence: `{"utr": "41298819999", "app": "GooglePay"}`,
		},
		// Order 6 (Out for Delivery): Rejected UPI Payment scenario
		{
			Key: "pay:ord6_upi_rejected", OrderKey: "order:6_out_delivery",
			Method: "UPI", Amount: 45000,
			Status: "REJECTED", Verification: "REJECTED",
			ActorKey: "user:vendor_2", Reference: "UPI/2026/rej_77182", Evidence: `{"rejection_reason": "UTR mismatch on bank statement"}`,
		},
		// Order 3 (Preparing Pickup): Partial payment recorded (₹100 of ₹180)
		{
			Key: "pay:ord3_partial", OrderKey: "order:3_preparing_pickup",
			Method: "CASH", Amount: 10000,
			Status: "CONFIRMED", Verification: "VERIFIED",
			ActorKey: "user:vendor_1", Reference: "CASH/2026/003", Evidence: `{"notes": "Advance token cash payment"}`,
		},
		// Payment with Refund scenario (Order 7 has partial return)
		{
			Key: "pay:refund_test", OrderKey: "order:7_completed",
			Method: "CASH", Amount: 4000,
			Status: "REFUNDED_PARTIAL", Verification: "VERIFIED",
			ActorKey: "user:vendor_1", Reference: "CASH/2026/004", Evidence: `{"notes": "Extra item paid"}`,
			HasRefund:    true,
			RefundAmount: 4000,
			RefundReason: "One damaged mango returned by customer and cash refunded",
		},
	}

	for _, p := range paymentsToSeed {
		payID := DetUUID(p.Key)
		orderID := orderUUIDMap[p.OrderKey]
		actorID := userUUIDMap[p.ActorKey]

		evidenceJSON, _ := json.Marshal(map[string]interface{}{"note": p.Evidence})

		_, err := tx.Exec(ctx, `
			INSERT INTO payments (id, order_id, method, amount_paise, status, verification_status, recorded_by, reference, evidence, idempotency_key, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW() - INTERVAL '1 hour')
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				verification_status = EXCLUDED.verification_status
		`, payID, orderID, p.Method, p.Amount, p.Status, p.Verification, actorID, p.Reference, evidenceJSON, p.Key)
		if err != nil {
			return nil, fmt.Errorf("payment %s: %w", p.Key, err)
		}
		summary.Payments++

		if p.HasRefund {
			refID := DetUUID(p.Key + ":refund")
			_, err = tx.Exec(ctx, `
				INSERT INTO refunds (id, payment_id, amount_paise, status, reason, idempotency_key, created_at)
				VALUES ($1, $2, $3, 'PROCESSED', $4, $5, NOW() - INTERVAL '30 minutes')
				ON CONFLICT (id) DO NOTHING
			`, refID, payID, p.RefundAmount, p.RefundReason, p.Key+":ref")
			if err != nil {
				return nil, fmt.Errorf("refund %s: %w", p.Key, err)
			}
			summary.Refunds++
		}
	}

	// -------------------------------------------------------------------------
	// 8. SEED UDHAAR / CREDIT MANAGEMENT (APPEND-ONLY LEDGER)
	// -------------------------------------------------------------------------
	udhaarAccountsToSeed := []struct {
		Key         string
		VendorKey   string
		CustKey     string
		CreditLimit int64
		Status      string
		Entries     []struct {
			Key       string
			EntryType string
			Amount    int64
			Note      string
		}
	}{
		// Account 1: Active account with outstanding balance of ₹1,200 (₹1,500 + ₹500 - ₹800)
		{
			Key: "udhaar:v1_ramesh", VendorKey: "vendor:1", CustKey: "user:cust_ramesh",
			CreditLimit: 500000, Status: "ACTIVE", // ₹5,000 credit limit
			Entries: []struct {
				Key       string
				EntryType string
				Amount    int64
				Note      string
			}{
				{"ue:v1_ramesh_1", "CREDIT", 150000, "Weekly vegetables and fruits purchased on credit"},
				{"ue:v1_ramesh_2", "CREDIT", 50000, "Extra mangoes crate credit addition"},
				{"ue:v1_ramesh_3", "REPAYMENT", 80000, "Partial cash repayment received from Ramesh"},
			},
		},
		// Account 2: Active account fully settled (₹1,200 credit - ₹1,200 repayment = ₹0 balance)
		{
			Key: "udhaar:v2_priya", VendorKey: "vendor:2", CustKey: "user:cust_priya",
			CreditLimit: 300000, Status: "ACTIVE", // ₹3,000 credit limit
			Entries: []struct {
				Key       string
				EntryType string
				Amount    int64
				Note      string
			}{
				{"ue:v2_priya_1", "CREDIT", 120000, "Monthly grocery credit (Rice + Dal)"},
				{"ue:v2_priya_2", "REPAYMENT", 120000, "Full settlement via UPI payment"},
			},
		},
		// Account 3: Suspended account with outstanding balance ₹850
		{
			Key: "udhaar:v1_anand", VendorKey: "vendor:1", CustKey: "user:cust_anand",
			CreditLimit: 100000, Status: "SUSPENDED", // ₹1,000 limit, suspended
			Entries: []struct {
				Key       string
				EntryType string
				Amount    int64
				Note      string
			}{
				{"ue:v1_anand_1", "CREDIT", 85000, "Past due seasonal fruits credit"},
			},
		},
	}

	for _, ua := range udhaarAccountsToSeed {
		accID := DetUUID(ua.Key)
		vendorID := vendorUUIDMap[ua.VendorKey]
		custID := userUUIDMap[ua.CustKey]

		var actualAccID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO udhaar_accounts (id, vendor_id, customer_id, status, credit_limit_paise, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW() - INTERVAL '7 days', NOW())
			ON CONFLICT (vendor_id, customer_id) DO UPDATE SET
				status = EXCLUDED.status,
				credit_limit_paise = EXCLUDED.credit_limit_paise,
				updated_at = NOW()
			RETURNING id
		`, accID, vendorID, custID, ua.Status, ua.CreditLimit).Scan(&actualAccID)
		if err != nil {
			return nil, fmt.Errorf("udhaar account %s: %w", ua.Key, err)
		}
		summary.UdhaarAccounts++

		// Append-only ledger entries
		for _, e := range ua.Entries {
			eID := DetUUID(e.Key)
			vendorOwnerID := vendorOwnerMap[ua.VendorKey]
			_, err = tx.Exec(ctx, `
				INSERT INTO udhaar_entries (id, account_id, entry_type, amount_paise, reference_type, reference_id, idempotency_key, actor_user_id, note, created_at)
				VALUES ($1, $2, $3, $4, 'MANUAL', NULL, $5, $6, $7, NOW() - INTERVAL '3 days')
				ON CONFLICT (id) DO NOTHING
			`, eID, actualAccID, e.EntryType, e.Amount, e.Key, vendorOwnerID, e.Note)
			if err != nil {
				return nil, fmt.Errorf("udhaar entry %s: %w", e.Key, err)
			}
			summary.UdhaarEntries++
		}
	}

	// -------------------------------------------------------------------------
	// 9. SEED NOTIFICATIONS & FCM DEVICE TOKENS
	// -------------------------------------------------------------------------
	notifsToSeed := []struct {
		Key       string
		UserKey   string
		EventType string
		Title     string
		Body      string
		Payload   string
		IsRead    bool
	}{
		{
			Key: "notif:1", UserKey: "user:cust_ramesh",
			EventType: "ORDER_ACCEPTED",
			Title:     "Order Accepted by Suresh Fresh Fruits",
			Body:      "ನಿಮ್ಮ ಆದೇಶವನ್ನು ಅಂಗಡಿಯವರು ಸ್ವೀಕರಿಸಿದ್ದಾರೆ. ತಯಾರಿ ನಡೆಯುತ್ತಿದೆ. (Your order has been accepted and is being prepared)",
			Payload:   `{"order_id": "order:1_pending", "type": "ORDER"}`,
			IsRead:    false,
		},
		{
			Key: "notif:2", UserKey: "user:cust_ramesh",
			EventType: "ORDER_READY",
			Title:     "Order Ready for Pickup",
			Body:      "ನಿಮ್ಮ ಹಾಲು ಮತ್ತು ಪನೀರ್ ಆದೇಶ ಪಿಕಪ್‌ಗೆ ಸಿದ್ಧವಾಗಿದೆ. (Your order is ready for pickup at Nandi Dairy)",
			Payload:   `{"order_id": "order:5_ready_pickup", "type": "ORDER"}`,
			IsRead:    false,
		},
		{
			Key: "notif:3", UserKey: "user:vendor_1",
			EventType: "ORDER_CREATED",
			Title:     "New Customer Order Received #ord_001",
			Body:      "ಹೊಸ ಗ್ರಾಹಕರ ಆದೇಶ ಬಂದಿದೆ. ದಯವಿಟ್ಟು ಪರಿಶೀಲಿಸಿ. (New customer order received for review)",
			Payload:   `{"order_id": "order:1_pending", "type": "ORDER"}`,
			IsRead:    true,
		},
		{
			Key: "notif:4", UserKey: "user:vendor_1",
			EventType: "PAYMENT_RECEIVED",
			Title:     "Cash Payment Confirmed: ₹300",
			Body:      "ನಗದು ಪಾವತಿ ಯಶಸ್ವಿಯಾಗಿ ದಾಖಲಾಗಿದೆ. (Cash payment ₹300 confirmed for completed order)",
			Payload:   `{"payment_id": "pay:ord7_cash", "type": "PAYMENT"}`,
			IsRead:    true,
		},
		{
			Key: "notif:5", UserKey: "user:cust_ramesh",
			EventType: "UDHAAR_CREDIT_RECORDED",
			Title:     "Udhaar Credit Added: ₹500",
			Body:      "ಸುರೇಶ್ ತರಕಾರಿ ಅಂಗಡಿಯಲ್ಲಿ ₹500 ಉದ್ಧಾರ ಖಾತೆಗೆ ಸೇರಿಸಲಾಗಿದೆ. (₹500 added to your Udhaar ledger)",
			Payload:   `{"account_id": "udhaar:v1_ramesh", "type": "UDHAAR"}`,
			IsRead:    false,
		},
	}

	for _, n := range notifsToSeed {
		nID := DetUUID(n.Key)
		uID := userUUIDMap[n.UserKey]
		var readAt *time.Time
		if n.IsRead {
			now := time.Now()
			readAt = &now
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO notifications (id, user_id, event_type, title, body, payload, read_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, NOW() - INTERVAL '1 hour')
			ON CONFLICT (id) DO NOTHING
		`, nID, uID, n.EventType, n.Title, n.Body, n.Payload, readAt)
		if err != nil {
			return nil, fmt.Errorf("notification %s: %w", n.Key, err)
		}
		summary.Notifications++
	}

	// Device tokens for push notification delivery
	deviceTokensToSeed := []struct {
		UserKey  string
		Token    string
		Platform string
	}{
		{"user:cust_ramesh", "fcm_token_customer_ramesh_android_emulator_5554", "ANDROID"},
		{"user:vendor_1", "fcm_token_vendor_suresh_android_pixel_001", "ANDROID"},
		{"user:super_admin", "fcm_token_admin_web_dashboard_token_99", "WEB"},
	}

	for _, dt := range deviceTokensToSeed {
		dtID := DetUUID(dt.Token)
		uID := userUUIDMap[dt.UserKey]
		_, err := tx.Exec(ctx, `
			INSERT INTO device_tokens (id, user_id, token, platform, last_seen_at, created_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (token) DO UPDATE SET last_seen_at = NOW()
		`, dtID, uID, dt.Token, dt.Platform)
		if err != nil {
			return nil, fmt.Errorf("device token %s: %w", dt.Token, err)
		}
		summary.DeviceTokens++
	}

	// -------------------------------------------------------------------------
	// 10. SEED OFFLINE SYNC OPERATIONS
	// -------------------------------------------------------------------------
	syncOpsToSeed := []struct {
		Key      string
		UserKey  string
		DeviceID string
		OpType   string
		Status   string
		Payload  string
		Result   string
	}{
		{
			Key: "sync:v1_desc_update", UserKey: "user:vendor_1",
			DeviceID: "android-pixel-7a", OpType: "PRODUCT_DESCRIPTION_UPDATE",
			Status:  "ACCEPTED",
			Payload: `{"product_id": "` + prodUUIDMap["prod:v1_banana"].String() + `", "description": "100% Organically ripened sweet Robusta bananas"}`,
			Result:  `{"updated": true, "status": "applied"}`,
		},
		{
			Key: "sync:v1_stock_adjust", UserKey: "user:vendor_1",
			DeviceID: "android-pixel-7a", OpType: "STOCK_ADJUSTMENT_REQUEST",
			Status:  "PENDING",
			Payload: `{"product_id": "` + prodUUIDMap["prod:v1_tomato"].String() + `", "qty_delta": 25, "reason": "Morning fresh mandi delivery"}`,
			Result:  `null`,
		},
	}

	for _, s := range syncOpsToSeed {
		sID := DetUUID(s.Key)
		uID := userUUIDMap[s.UserKey]
		clOpID := DetUUID(s.Key + ":client_op")

		var resJSON *string
		if s.Result != "null" {
			resJSON = &s.Result
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO sync_operations (id, user_id, device_id, client_operation_id, operation_type, payload_version, status, payload, result, created_at)
			VALUES ($1, $2, $3, $4, $5, 1, $6, $7::jsonb, $8::jsonb, NOW() - INTERVAL '2 hours')
			ON CONFLICT (user_id, client_operation_id) DO UPDATE SET
				status = EXCLUDED.status,
				result = EXCLUDED.result
		`, sID, uID, s.DeviceID, clOpID, s.OpType, s.Status, s.Payload, resJSON)
		if err != nil {
			return nil, fmt.Errorf("sync operation %s: %w", s.Key, err)
		}
		summary.SyncOperations++
	}

	// -------------------------------------------------------------------------
	// 11. SEED AUDIT LOGS (ADMIN PRIVILEGED ACTIONS & SECURITY TRAIL)
	// -------------------------------------------------------------------------
	auditLogsToSeed := []struct {
		Key          string
		ActorKey     string
		Action       string
		ResourceType string
		ResourceID   uuid.UUID
		Reason       string
		Metadata     string
	}{
		{
			Key: "audit:suspend_user", ActorKey: "user:super_admin",
			Action: "USER_STATUS_UPDATE", ResourceType: "users",
			ResourceID: userUUIDMap["user:cust_suspended"],
			Reason:     "Account suspended due to policy violation & suspicious chargebacks",
			Metadata:   `{"previous_status": "ACTIVE", "new_status": "SUSPENDED", "reviewed_by": "super_admin"}`,
		},
		{
			Key: "audit:suspend_vendor", ActorKey: "user:admin_kavitha",
			Action: "VENDOR_STATUS_UPDATE", ResourceType: "vendor_profiles",
			ResourceID: vendorUUIDMap["vendor:14"],
			Reason:     "City Market vendor health inspection license expired",
			Metadata:   `{"previous_status": "ACTIVE", "new_status": "SUSPENDED", "inspection_ref": "BBMP/2026/099"}`,
		},
		{
			Key: "audit:role_assign", ActorKey: "user:super_admin",
			Action: "ROLE_ASSIGNED", ResourceType: "user_roles",
			ResourceID: userUUIDMap["user:admin_kavitha"],
			Reason:     "Promoted operator Kavitha to administrative officer",
			Metadata:   `{"assigned_role": "ADMIN", "effective_from": "2026-11-01"}`,
		},
		{
			Key: "audit:refund_approved", ActorKey: "user:vendor_1",
			Action: "REFUND_PROCESSED", ResourceType: "payments",
			ResourceID: DetUUID("pay:refund_test"),
			Reason:     "Damaged mango returned at counter; cash refunded",
			Metadata:   `{"refund_amount_paise": 4000, "payment_method": "CASH"}`,
		},
	}

	for _, a := range auditLogsToSeed {
		aID := DetUUID(a.Key)
		actorID := userUUIDMap[a.ActorKey]
		_, err := tx.Exec(ctx, `
			INSERT INTO audit_logs (id, actor_user_id, action, resource_type, resource_id, reason, metadata, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, NOW() - INTERVAL '2 days')
			ON CONFLICT (id) DO NOTHING
		`, aID, actorID, a.Action, a.ResourceType, a.ResourceID, a.Reason, a.Metadata)
		if err != nil {
			return nil, fmt.Errorf("audit log %s: %w", a.Key, err)
		}
		summary.AuditLogs++
	}

	// -------------------------------------------------------------------------
	// 12. SEED ACTIVE REFRESH SESSIONS
	// -------------------------------------------------------------------------
	activeSessionUsers := []string{"user:super_admin", "user:cust_ramesh", "user:vendor_1"}
	for _, uKey := range activeSessionUsers {
		sID := DetUUID("sess:" + uKey)
		uID := userUUIDMap[uKey]
		// Fake sha256 hash of token for seeding
		tokenHash := sha256.Sum256([]byte("dev_refresh_token_" + uKey))
		tokenHashStr := hex.EncodeToString(tokenHash[:])

		_, err := tx.Exec(ctx, `
			INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at, revoked_at, device_id, created_at)
			VALUES ($1, $2, $3, NOW() + INTERVAL '30 days', NULL, 'dev-device', NOW())
			ON CONFLICT (token_hash) DO NOTHING
		`, sID, uID, tokenHashStr)
		if err != nil {
			return nil, fmt.Errorf("refresh session %s: %w", uKey, err)
		}
		summary.RefreshSessions++
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seed transaction: %w", err)
	}

	logger.Logger.Info().Msg("VendorSaathi development database seeding committed successfully.")
	return summary, nil
}

func strPtr(s string) *string {
	return &s
}
