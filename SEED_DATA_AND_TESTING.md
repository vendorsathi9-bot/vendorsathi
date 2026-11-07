# VendorSaathi — Seed Data & Comprehensive Feature Verification Guide

This document provides a complete reference for the development database seeding system, the seeded datasets, and end-to-end verification workflows across the **Go Backend API**, the **Flutter Mobile Application**, and the **React Admin Dashboard**.

---

## 1. Executive Summary & Architecture Overview

VendorSaathi is a hyper-local commerce and credit platform tailored for Indian street and neighborhood vendors, kirana stores, and consumers.

### Development Location Anchor
All geo-spatial discovery testing is calibrated to **Bengaluru Majestic (City Railway Station / Kempegowda Bus Station)**:
- **Reference Latitude:** `12.9767`
- **Reference Longitude:** `77.5713`
- **Discovery Radius:** `10.0 km`
- **Haversine Implementation:** Pure SQL spherical distance calculation in `internal/modules/mobility/repository/mobility_repo.go`.

### Database Architecture & Populated Entities
The database seeding system (`internal/platform/seeder/seeder.go`) executes within PostgreSQL transactions, enforcing referential integrity, check constraints, and business invariants across all 30 database tables:

| Database Table | Seeded Count | Purpose & Invariants Verified |
|---|---|---|
| `users` | 31 | Auth credentials across Super Admin, Admin, Vendor, Customer, and Suspended users |
| `user_roles` | 32 | RBAC mapping for `SUPER_ADMIN`, `ADMIN`, `VENDOR`, and `CUSTOMER` |
| `roles` | 4 | System roles (`SUPER_ADMIN`, `ADMIN`, `VENDOR`, `CUSTOMER`) |
| `permissions` | 23 | Granular permissions (e.g., `orders:read`, `inventory:update`, etc.) |
| `role_permissions` | 57 | Default permission bindings per role |
| `vendor_profiles` | 20 | 13 active (<10km), 2 unapproved/suspended (<10km), 5 active (>10km) |
| `vendor_locations` | 20 | Exact coordinates around Bangalore Majestic |
| `business_hours` | 140 | Operating hours across all 7 days for every seeded vendor |
| `customer_profiles` | 7 | Kannada & English customer profiles with preferred language |
| `customer_addresses` | 7 | Residential and workplace delivery addresses |
| `categories` | 6 | Fruits & Veggies, Groceries, Bakery, Flowers, Meat, Street Food |
| `products` | 27 | Realistic regional products across all categories |
| `inventory` | 27 | In-stock, low-stock, and out-of-stock items with non-negative constraints |
| `inventory_movements` | 27 | Initial stock batch movements with audit actor references |
| `carts` | 1 | Single-vendor cart (Customer Ramesh @ Suresh Fresh Fruits) |
| `cart_items` | 3 | 3 distinct product items with line pricing |
| `orders` | 11 | Complete state machine coverage (`PENDING` through `COMPLETED`, `CANCELLED`, etc.) |
| `order_items` | 22 | Price snapshots and quantities for all orders |
| `order_status_history` | 34 | Complete state transition audit trail with timestamps |
| `order_cancellations` | 2 | Cancellation requests (1 approved with refund, 1 pending review) |
| `payments` | 7 | Payment records (Cash, UPI confirmed, UPI pending verification, UPI rejected) |
| `refunds` | 1 | UPI reversal linked to approved cancellation |
| `udhaar_accounts` | 3 | Append-only credit accounts (Active with balance, Fully settled, Suspended) |
| `udhaar_entries` | 6 | Strict double-entry ledger records (`CREDIT`, `REPAYMENT`) |
| `notifications` | 5 | In-app alerts for order statuses and Udhaar updates |
| `device_tokens` | 3 | Mobile push registration tokens (FCM) |
| `sync_operations` | 2 | Client offline-first sync operational logs |
| `audit_logs` | 4 | System-level admin action audit trails |
| `refresh_sessions` | 40+ | Active JWT refresh tokens for fast login |
| `schema_migrations` | 5 | Migration versions tracked up to `005_business_hours_updated_at` |

---

## 2. Seed Execution & CLI Commands

The database seeding mechanism is built into the Go backend using standard package architecture and dependency injection.

### A. Non-Destructive Idempotent Seeding
Executes all inserts with `ON CONFLICT DO UPDATE` or key-checks. Safe to run repeatedly at any time without creating duplicates:

```bash
# Using main API CLI
go run ./cmd/api seed

# OR using dedicated seeder CLI
go run ./cmd/seed
```

### B. Full Database Reset & Re-Seed
If you require a completely fresh database state, use the `--reset` flag. For security, this command **strictly refuses** to run in production and requires the environment confirmation `CONFIRM_DB_RESET=yes`:

```bash
CONFIRM_DB_RESET=yes go run ./cmd/api seed --reset
# OR
CONFIRM_DB_RESET=yes go run ./cmd/seed --reset
```

---

## 3. Test Credentials & Role Accounts

All seeded passwords use standard BCrypt hashing compatible with the Go auth handler (`cost = 10` in dev).

| Role | Phone Number | Password | Profile Name / Identifier | Intended Testing Purpose |
|---|---|---|---|---|
| **SUPER_ADMIN** | `+919876543210` | `AdminPassword123!` | Platform Super Administrator | Admin dashboard overview, order governance, role audit |
| **ADMIN** | `+919876543211` | `AdminPassword123!` | Kavitha Operations Lead | Admin dashboard management & user status modification |
| **ADMIN (Suspended)** | `+919876543219` | `AdminPassword123!` | Suspended Admin Account | Auth negative testing: verifies `account is not active` |
| **VENDOR** | `+919876500001` | `VendorPassword123!` | Suresh Fresh Fruits (ಸುರೇಶ್) | 0.08 km away. Pending & preparing orders, catalog, Udhaar |
| **VENDOR** | `+919876500002` | `VendorPassword123!` | Chickpet Sri Krishna Provisions | 0.87 km away. Groceries, Udhaar ledger, accepted order |
| **VENDOR** | `+919876500003` | `VendorPassword123!` | Gandhi Nagar Organic Greens | 0.57 km away. Delivery order, out for delivery flow |
| **VENDOR** | `+919876500004` | `VendorPassword123!` | Malleshwaram Nandi Dairy | 2.99 km away. Completed orders, full cash settlement |
| **VENDOR** | `+919876500005` | `VendorPassword123!` | Rajajinagar Iyengar Bakery | 2.53 km away. Bakery goods, low stock scenarios |
| **VENDOR (Suspended)**| `+919876500014` | `VendorPassword123!` | Majestic Street Corn Stall | 0.45 km away. Discovery exclusion: verifies unapproved vendors hidden |
| **VENDOR (Outside 10km)**| `+919876500016` | `VendorPassword123!` | Whitefield Super Provision | 16.8 km away. Discovery exclusion: verifies distance filter |
| **CUSTOMER** | `+919888877777` | `CustomerPassword123!` | Ramesh Gowda (ರಮೇಶ್ ಗೌಡ) | Active cart, 4 orders in various states, Udhaar debtor |
| **CUSTOMER** | `+919888811111` | `CustomerPassword123!` | Priya Sharma (ಪ್ರಿಯಾ ಶರ್ಮಾ) | Placed pickup order, Udhaar settled account |
| **CUSTOMER** | `+919888822222` | `CustomerPassword123!` | Anand Kumar (ಆನಂದ್ ಕುಮಾರ್) | Placed delivery order, suspended Udhaar account |
| **CUSTOMER (Suspended)**| `+919888899999` | `CustomerPassword123!` | Suspended Customer Account | Auth negative testing: verifies blocked login |

---

## 4. Comprehensive Feature Verification Matrix

| Feature | Seeded Data | Verification Steps | Expected Result |
|---|---|---|---|
| **Authentication — Valid Login** | Super Admin (`+919876543210`), Vendor (`+919876500001`), Customer (`+919888877777`) | `POST /api/v1/auth/login` with phone and correct password | HTTP 200 with `access_token`, `refresh_token`, and user metadata |
| **Authentication — Inactive Account Guard** | Suspended Admin (`+919876543219`), Suspended Customer (`+919888899999`) | `POST /api/v1/auth/login` with correct password | HTTP 403 Forbidden with error `"account is not active"` |
| **Authentication — Incorrect Password** | Any seeded phone with password `WrongPassword!` | `POST /api/v1/auth/login` | HTTP 401 Unauthorized with error `"invalid credentials"` |
| **Token Refresh** | Stored `refresh_sessions` records | `POST /api/v1/auth/refresh` with valid refresh token | HTTP 200 with new `access_token` and rotated `refresh_token` |
| **Nearby Vendor Discovery (<10km)** | 13 active vendors located between 0.08 km and 7.59 km from Majestic (`12.9767, 77.5713`) | `GET /api/v1/mobility/nearby-vendors?lat=12.9767&lng=77.5713&radius_km=10.0` | HTTP 200 returning array of exactly 13 active vendors sorted ascending by distance |
| **Radius Filtering Verification** | 5 vendors located between 10.6 km and 20.3 km (e.g., Whitefield, Electronic City, Yelahanka) | `GET /api/v1/mobility/nearby-vendors?lat=12.9767&lng=77.5713&radius_km=1.0` | HTTP 200 returning only 3 vendors within 1.0 km; vendors beyond 1.0 km are excluded |
| **Unapproved Vendor Exclusion** | Stall 14 (`+919876500014`, status `SUSPENDED`) and Push Cart (`+919876500015`, status `PENDING`) located 0.45 km away | Query `/api/v1/mobility/nearby-vendors` with 10.0 km radius | Vendors with status `SUSPENDED` or `PENDING` are omitted from discovery results |
| **Product Catalogue by Vendor** | 8 products assigned to Suresh Fresh Fruits (Bananas, Tomatoes, Onions, Potatoes, Apples, etc.) | `GET /api/v1/products?vendor_id=<suresh_id>` or `GET /api/v1/products` (as vendor) | HTTP 200 returning all 8 products with accurate prices, SKU, and categories |
| **Inventory Tracking & Status** | Robusta Banana (In-stock, qty 85), Country Tomato (Low-stock, qty 8), Mahabaleshwar Strawberry (Out-of-stock, qty 0) | `GET /api/v1/inventory` as Suresh or query inventory records directly | Quantities match seeded levels: Banana 85, Tomato 8, Strawberry 0 |
| **Inventory Stock Adjustment** | Seeded product Robusta Banana | `POST /api/v1/inventory/adjust` with `product_id`, `qty_delta: 5`, `reason: "Restock"` | HTTP 200, stock increments from 85 to 90, entry appended to `inventory_movements` |
| **Customer Cart Operations** | Ramesh (`+919888877777`) has active cart with 3 items from Suresh Fresh Fruits | `GET /api/v1/carts/me` with Ramesh's Bearer token | HTTP 200 returning cart with items: Robusta Banana (2), Country Tomato (3), Nagpur Orange (1) |
| **Single-Vendor Cart Invariant** | Ramesh cart tied to Suresh Fresh Fruits | `POST /api/v1/carts/items` with a product from Iyengar Bakery | HTTP 400 Bad Request: Cart cannot contain items from multiple vendors simultaneously |
| **Order Checkout Flow** | Ramesh's active cart | `POST /api/v1/orders` with `fulfilment_type: "PICKUP"` | HTTP 201 Created: Order created with status `PENDING`, items copied to `order_items`, cart cleared |
| **Vendor Order Management** | 6 orders assigned to Suresh Fresh Fruits across `PENDING`, `ACCEPTED`, `PREPARING`, `COMPLETED`, `REJECTED` | `GET /api/v1/orders` with Suresh's Bearer token | HTTP 200 returning vendor's 6 orders with customer IDs and formatted item lines |
| **Order State Machine: Accept** | Order #1 (`PENDING`) | `POST /api/v1/orders/<order_1_id>/accept` with Suresh's Bearer token | HTTP 200: Status updates to `ACCEPTED`, entry recorded in `order_status_history` |
| **Order State Machine: Prepare** | Order #2 (`ACCEPTED`) | `POST /api/v1/orders/<order_2_id>/prepare` | HTTP 200: Status updates to `PREPARING` |
| **Order State Machine: Ready for Pickup**| Order #3 (`PREPARING`, PICKUP) | `POST /api/v1/orders/<order_3_id>/ready` | HTTP 200: Status updates to `READY_FOR_PICKUP` |
| **Order State Machine: Out for Delivery**| Order #5 (`PREPARING`, DELIVERY) | `POST /api/v1/orders/<order_5_id>/out-for-delivery` | HTTP 200: Status updates to `OUT_FOR_DELIVERY` |
| **Order State Machine: Complete** | Order #4 (`READY_FOR_PICKUP`) | `POST /api/v1/orders/<order_4_id>/complete` | HTTP 200: Status updates to `COMPLETED` |
| **Order State Machine: Invalid Transition**| Order #6 (`COMPLETED`) | `POST /api/v1/orders/<order_6_id>/cancel` | HTTP 400/422 Unprocessable: Completed orders cannot transition to `CANCELLED` |
| **Cancellation & Refund Flow** | Order #9 (`CANCELLED`, approved) and Order #11 (`CANCELLATION_REQUESTED`) | `GET /api/v1/orders` checking cancellation details | Approved cancellation has record in `order_cancellations` and `refunds` (UPI reversal) |
| **Payments — Cash Settlement** | Order #6 (COMPLETED, ₹400.00 cash payment) | Query `payments` table or order payment details | Payment method `CASH`, status `CONFIRMED`, amount 40000 paise |
| **Payments — UPI Confirmed** | Order #2 (ACCEPTED, ₹220.00 UPI payment) | Query `payments` table for UPI reference | Payment method `UPI`, status `CONFIRMED`, transaction ref `UPI-TXN-20261107-001` |
| **Payments — UPI Verification Pending**| Order #7 (OUT_FOR_DELIVERY, ₹350.00 UPI) | Query `payments` table | Payment method `UPI`, status `PENDING`, awaiting vendor verification |
| **Udhaar Credit Account Viewing** | Ramesh Gowda has ₹1,200.00 outstanding Udhaar balance with Suresh Fresh Fruits | `GET /api/v1/udhaar/accounts` as Suresh | HTTP 200 returning Ramesh's account with `balance_paise: 120000`, status `ACTIVE` |
| **Udhaar Append-Only Ledger** | 6 historical transactions across 3 customer accounts (`CREDIT` and `REPAYMENT`) | `GET /api/v1/udhaar/accounts/<id>/entries` | HTTP 200 returning sequential ledger entries. Current balance matches sum of credits minus repayments |
| **Customer Notifications** | Ramesh has 3 unread notifications (Order Accepted, Order Ready, Udhaar Credit Added) | `GET /api/v1/notifications` with Ramesh's Bearer token | HTTP 200 returning notifications list with Kannada titles and localized bodies |
| **Admin Platform Overview** | Seeded system-wide records | `GET /api/v1/analytics/platform/overview` with Admin token | HTTP 200 returning aggregates: 32 users, 20 vendors (18 active), 11 orders, ₹800.00 gross sales |
| **Admin All Orders Listing** | 11 orders platform-wide | `GET /api/v1/orders?limit=50` with Admin token | HTTP 200 returning all 11 orders across all vendors and customers |
| **Admin User Management** | Seeded users with statuses | `PATCH /api/v1/users/<id>/status` with `{ "status": "SUSPENDED" }` | HTTP 200, user status updated, entry written to `audit_logs` |
| **Admin Audit Trail** | 4 seeded audit entries (`VENDOR_STATUS_UPDATE`, `REFUND_PROCESSED`, etc.) | `GET /api/v1/audit-logs?limit=20` with Admin token | HTTP 200 returning chronologically sorted audit log entries with actor user IDs |

---

## 5. Walkthrough: Flutter Mobile Application Testing

### A. Customer Experience (Discover, Cart, My Orders)
1. **Launch the App:** Open the mobile app on the emulator (`emulator-5554`).
2. **Login as Customer:**
   - Phone: `+919888877777`
   - Password: `CustomerPassword123!`
   - Tap **Sign In (ಲಾಗಿನ್)**.
3. **Discover Tab (ಹತ್ತಿರದ ವ್ಯಾಪಾರಿಗಳು):**
   - The Discover screen will immediately display all active vendors within 10 km sorted by proximity:
     - *Suresh Fresh Fruits & Vegetables* (0.1 km away)
     - *Gandhi Nagar Organic Greens* (0.6 km away)
     - *Chickpet Sri Krishna Kirana* (0.9 km away)
     - *Rajajinagar Iyengar Bakery* (2.5 km away)
     - *Malleshwaram Nandi Dairy* (3.0 km away)
     - ...up to 13 vendors within the radius.
   - Note that unapproved vendors (e.g., Majestic Street Corn Stall) and distant vendors (Whitefield, Electronic City) are correctly omitted.
4. **Cart Tab (ಬುಟ್ಟಿ):**
   - Tap the **Cart** tab.
   - Ramesh's pre-seeded cart contains 3 items from Suresh Fresh Fruits:
     - *Robusta Banana (2 kg)*
     - *Country Tomato (3 kg)*
     - *Nagpur Orange (1 kg)*
   - Tap **Checkout (Pick-up)** to place the order. The order is submitted to the backend, the cart empties, and navigation switches to **My Orders**.
5. **My Orders Tab (ನನ್ನ ಆರ್ಡರ್‌ಗಳು):**
   - View seeded orders in various operational states (`PENDING`, `READY_FOR_PICKUP`, `COMPLETED`, `CANCELLATION_REQUESTED`).

---

### B. Vendor Experience (Summary, Orders, Products, Udhaar)
1. **Logout & Switch Account:**
   - Tap the Logout icon in the upper right.
2. **Login as Vendor:**
   - Phone: `+919876500001` (Suresh - Fruit & Veg Vendor)
   - Password: `VendorPassword123!`
   - Tap **Sign In (ಲಾಗಿನ್)**.
3. **Tab 1 — Summary (ವ್ಯಾಪಾರಿ ಮುಖಪುಟ):**
   - Displays real-time financial metrics:
     - **Completed Sales:** `₹300.00` (6 orders completed)
     - **Customer Udhaar (ಉಧಾರಿ ಬಾಕಿ):** `₹1,200.00` outstanding across customers.
4. **Tab 2 — Orders (ಆರ್ಡರ್‌ಗಳು):**
   - Displays real-time order action cards:
     - **Order #a19579f7 (`PENDING`, ₹140.00):** Has **[Accept]** and **[Reject]** buttons.
     - **Order #f4b6c45a (`ACCEPTED`, ₹220.00):** Has **[Start Preparing]** button.
     - **Order #02e670af (`PREPARING`, ₹180.00):** Has **[Ready for Pickup]** button.
   - Tap any action button to execute an immediate state transition via the API.
5. **Tab 3 — Products (ಉತ್ಪನ್ನಗಳು):**
   - Displays Suresh's 8 products (Robusta Banana, Country Tomato, Bellary Onion, Shimla Apple, etc.) with real-time inventory counts and quick stock adjustment controls (`+1`, `-1`).
6. **Tab 4 — Udhaar Ledger (ಉಧಾರಿ ಖಾತೆ):**
   - Displays customer credit accounts:
     - **Ramesh Gowda:** `₹1,200.00` (`ACTIVE`)
     - **Anand Kumar:** `₹850.00` (`SUSPENDED`)

---

## 6. Walkthrough: React Admin Dashboard Testing

1. **Access Web Application:**
   - Open browser at `http://localhost:3000`.
2. **Sign In:**
   - Phone: `+919876543210`
   - Password: `AdminPassword123!`
3. **Dashboard (Platform Overview):**
   - Real-time aggregate metric cards populated from `/analytics/platform/overview`:
     - **Active Vendors:** 18 (out of 20 total registered)
     - **Total Customers:** 8
     - **Total Orders:** 11 (2 Completed)
     - **Gross Sales:** `₹800.00`
     - **Total Collections:** `₹900.00`
4. **Vendor Management (`/vendors`):**
   - Tabular view of all registered vendors with status indicators (`ACTIVE`, `PENDING`, `SUSPENDED`).
   - Click **Suspend** on any active vendor to trigger `PATCH /users/:id/status`, instantly writing an audit entry.
5. **Customer Management (`/customers`):**
   - Tabular view of all registered customers with phone numbers and registration dates.
6. **Order Governance (`/orders`):**
   - Comprehensive audit table of all 11 orders across the entire platform with status badges (`PENDING`, `PREPARING`, `OUT_FOR_DELIVERY`, `COMPLETED`, `CANCELLED`, etc.) and snapshot prices.
7. **Audit Logs (`/audit`):**
   - Chronological system activity records verifying actions such as `ROLE_ASSIGNED`, `USER_STATUS_UPDATE`, `REFUND_PROCESSED`, and `VENDOR_STATUS_UPDATE`.

---

## 7. Deferred / Unimplemented Features

During inspection and verification, the following items were identified as external mock dependencies or planned roadmap capabilities not present in the current codebase:
1. **Third-Party Payment Gateways (Razorpay/PayTM Webhooks):**
   - UPI payments currently rely on in-app confirmation and vendor verification. Automated bank webhook callbacks are not implemented.
2. **SMS OTP Gateway (Twilio/Gupshup):**
   - Authentication utilizes BCrypt phone-and-password auth. OTP verification tables are stubbed for future cellular SMS delivery.
3. **Live GPS Rider Tracking:**
   - Orders support `OUT_FOR_DELIVERY` status, but real-time WebSocket rider coordinate streaming is not implemented.
4. **Voice-Assisted Commerce (Bhashini AI / Kannada STT):**
   - Kannada bilingual labels and descriptions are fully seeded and rendered in the UI; automated voice transcription microservices are deferred.

---

## 8. Exact Service Restart Commands

### Start or Restart PostgreSQL Container
```bash
docker start vendorsaathi-postgres
```

### Start Backend API Server
```bash
# Compile and start daemon
go build -o ./bin/vendorsaathi ./cmd/api
./bin/vendorsaathi
```

### Start React Admin Web Dashboard
```bash
cd admin
npm run dev
# Accessible at http://localhost:3000
```

### Start Flutter Mobile Application
```bash
cd mobile
flutter run -d emulator-5554
```

### Send Deterministic GPS Fix to Emulator (Majestic Coordinates)
```bash
adb -s emulator-5554 emu geo fix 77.5713 12.9767
```
