# VendorSaathi Development Progress

**Last Updated:** 2026-11-06

## Summary of Completion

VendorSaathi is a local-commerce platform that digitizes vendor profiles, products, inventory, orders, fulfilment, payments, Udhaar (credit ledger), offline outbox synchronization, mobile apps, and administrative controls.

---

## Phase 1: Backend Foundation — ✅ COMPLETED
- [x] Go module setup with modular monolith directory structure
- [x] Config management with environment variable overrides and validations
- [x] PostgreSQL connection pool (`pgxpool`) with context propagation and timeouts
- [x] Structured logger with `zerolog` and request ID tracing
- [x] Authentication platform: Argon2id password hashing, JWT access + refresh token lifecycle
- [x] Standardized error envelopes and cursor pagination helpers (`internal/platform/httpx`)
- [x] Middleware: Request ID, recovery, CORS, JWT auth, RBAC role gating, permission checks, in-memory rate limiting
- [x] Migrations:
  - `001_initial_schema`: roles, permissions, users, user_roles, refresh_sessions
  - `002_vendor_customer_product`: vendor profiles, locations, hours, customer profiles, addresses, categories, products, inventory
  - `003_carts_orders`: carts, cart items, orders, order items, status history, order cancellations
  - `004_payments_udhaar_notifications`: payments, refunds, udhaar accounts, udhaar entries, notifications, device tokens, sync operations, audit logs

---

## Phase 2: Core Backend Modules — ✅ COMPLETED
All 14 core backend modules implemented with model, repository, service, and handler layers:
- [x] **Auth & Users Module**: Login, register, refresh tokens, profile, password changes, account status controls
- [x] **Vendors Module**: Profile CRUD, operating locations, weekly business hours, availability toggle
- [x] **Customers Module**: Customer profiles, delivery & pickup addresses
- [x] **Products & Categories Module**: Catalogue management, price validation in integer paise, categories taxonomy, archiving
- [x] **Inventory Module**: Atomic stock adjustments, movement history, optimistic locking, non-negative invariants
- [x] **Carts Module**: Single-vendor cart enforcement, item upsert, quantity modifications, cart clearing
- [x] **Orders Module**:
  - Deterministic row-locking stock reservation algorithm
  - Atomic checkout transaction in single database transaction
  - Complete state machine: PENDING → ACCEPTED → PREPARING → READY_FOR_PICKUP / OUT_FOR_DELIVERY → COMPLETED
  - Cancellation workflow (request cancellation, vendor approve/decline, automatic stock release on rejection/cancellation)
  - Full audit status history tracking
- [x] **Payments & Refunds Module**:
  - Cash and manually verified UPI records
  - Overpayment validation against order total
  - Split payment support (multiple payment records per order)
  - Refund records bounded by payment amount
- [x] **Udhaar Module**:
  - Vendor-managed customer credit accounts
  - Append-only ledger entries (`CREDIT`, `REPAYMENT`, `ADJUSTMENT`)
  - Concurrency safety with `SELECT ... FOR UPDATE` row locks
  - Balance derivation from ledger entries (no direct balance editing)
  - Repayment limit checks (cannot repay more than outstanding balance)
- [x] **Notifications Module**: In-app notifications, unread count tracking, device tokens registration/revocation
- [x] **Sync Module**: Bounded offline operation intake for allowlisted types (`PRODUCT_DESCRIPTION_UPDATE`, `STOCK_ADJUSTMENT_REQUEST`), idempotency deduplication by `(user_id, client_operation_id)`
- [x] **Mobility Module**: Haversine distance calculation and bounding-box spatial filtering for nearby vendor discovery
- [x] **Analytics Module**: Vendor sales summaries (gross sales, completed orders, outstanding Udhaar, confirmed collections) and admin platform overview
- [x] **Audit Module**: Append-only administrative and sensitive action audit trail

---

## Phase 3: Flutter Mobile Application — ✅ COMPLETED
Located in `/mobile`:
- [x] `pubspec.yaml` with Riverpod, SQLite (`sqflite`), secure storage, HTTP, Geolocator, Intl
- [x] Clean Material 3 design theme with Indian localized palette and high accessibility
- [x] Dual-language localization framework supporting English and Kannada (`kn`)
- [x] Typed `ApiClient` with error envelopes, timeouts, and automatic Bearer JWT handling
- [x] Device-local SQLite database (`AppDatabase`) with cached products, inventory snapshots, and outbox operations
- [x] Bounded offline `SyncRunner` executing pending outbox queue with the Go server
- [x] Role-based experiences:
  - **Vendor Experience**: Dashboard sales summary, real-time stock adjustments, product catalogue, order lifecycle controls (Accept, Reject, Prepare, Ready, Complete), Udhaar customer accounts
  - **Customer Experience**: Location-aware nearby vendor discovery, product browsing, cart management with single-vendor policy, checkout (Pickup / Delivery), order history tracking

---

## Phase 4: React Admin Dashboard — ✅ COMPLETED
Located in `/admin`:
- [x] Built with React 18, TypeScript, and Vite
- [x] Production build verified with `npm run build`
- [x] Clean, responsive administrative interface with Lucide icons
- [x] Core capabilities:
  - Admin authentication with JWT session persistence
  - Dashboard overview with real-time aggregate metrics (active vendors, gross sales, collections, orders)
  - Vendor management & account suspension/activation controls
  - Customer oversight
  - Order monitoring with full item snapshots & fulfilment types
  - Platform audit trail with immutable timestamps and actor IDs

---

## Phase 5: Integration & API Specifications — ✅ COMPLETED
- [x] Complete OpenAPI 3.0 specification in `api/openapi.yaml`
- [x] All backend handlers wired in `internal/app/app.go`
- [x] Consistent integer paise currency representation across DB, API, Flutter, and React
- [x] Standardized error contracts and HTTP status codes across the stack

---

## Phase 6: Testing & Production Readiness — ✅ COMPLETED
- [x] Unit test suite in `tests/unit_test.go` passing 100%:
  - Argon2id password hashing and verification
  - JWT token pair generation, claim parsing, role extraction
  - Order state machine transitions and invalid transition rejections
  - Financial invariants: integer paise calculations, overpayment rejection, Udhaar balance bounds
  - Haversine distance accuracy
- [x] `go build -o bin/vendorsaathi ./cmd/api` compiles cleanly
- [x] `go build -o bin/migrate ./cmd/migrate` compiles cleanly
- [x] `go vet ./...` reports zero issues
- [x] Dockerfile and Docker Compose configurations present for PostgreSQL, MinIO, and Go API

---

## Phase 7: Runtime Blockers Fixed — ✅ COMPLETED (2026-11-06)

### Files Changed

| File | Change |
|------|--------|
| `internal/config/config.go` | Fixed nested env-var binding (JWT_SIGNING_KEY → JWT.SIGNING_KEY), added explicit `BindEnv` for all required keys, added JSON/comma-separated CORS_ORIGINS parsing, improved error messages |
| `internal/platform/database/postgres.go` | Fixed `GetPgxTx` return type (`pgx.Tx` not `*pgxpool.Tx`), removed broken `GetTx`, added `pgx` import |
| `internal/platform/migration/runner.go` | **NEW** — Versioned SQL migration runner using pgx directly (no new library) |
| `cmd/migrate/main.go` | **NEW** — `go run ./cmd/migrate up / status` CLI |
| `internal/modules/auth/repository/user_repo.go` | Removed broken `WithTx()` that created an empty `pgxpool.Pool` |
| `internal/modules/products/handler/product_handler.go` | Fixed Gin route conflict: `c.Param("categoryId")` → `c.Param("id")` |
| `internal/app/app.go` | Fixed Gin route conflict: `/:categoryId/products` → `/:id/products` |
| `docker-compose.yml` | Removed obsolete `version:` attribute; made `config.yaml` volume mount optional (commented out); switched MinIO to `bitnami/minio:latest` (reliable image); updated MinIO volume path |
| `config.yaml` | Created from `config.yaml.example` with `localhost:5433` for host-based dev |

### Problems Fixed

1. **Viper nested env-var binding** — `JWT_SIGNING_KEY`, `DATABASE_URL`, etc. were silently ignored
   because Viper's `AutomaticEnv` does not automatically map flat `FOO_BAR` env vars to nested
   `FOO.BAR` config keys without an explicit `BindEnv`. Added `BindEnv` for every required key.

2. **CORS_ORIGINS JSON parsing** — When set as `'["...","..."]'` via Docker Compose env, Viper
   does not deserialize JSON arrays. Added explicit `json.Unmarshal` / comma-split fallback.

3. **Migration runner missing** — No migration runner existed; the application had no way to apply
   SQL files at startup or on demand. Implemented `internal/platform/migration.Runner` and
   `cmd/migrate` without adding any new library dependency.

4. **`UserRepository.WithTx`** — Created `&pgxpool.Pool{}` (an empty zero-value struct) instead of
   using the supplied `pgx.Tx`. Method was unused; removed entirely.

5. **`GetPgxTx` wrong type** — Returned `*pgxpool.Tx` (which is not a type exported by pgxpool);
   corrected to `pgx.Tx` (the interface stored by `TxManager.RunInTx`).

6. **Docker Compose `config.yaml` mount** — The bind mount required `config.yaml` to exist on the
   host even though all required values are already provided as environment variables in the same
   compose file. Mount is now commented out; the API works from env vars alone in Docker.

7. **MinIO image** — `minio/minio:latest` is frequently rate-limited / unavailable. Switched to
   `bitnami/minio:latest` with the correct data-volume path (`/bitnami/minio/data`).

8. **Gin route conflict** — `/api/v1/categories/:id` and `/api/v1/categories/:categoryId/products`
   used different wildcard names in the same radix-tree prefix, causing a panic at startup.
   Unified to `/:id/products` and updated the handler accordingly.

---

## Migration Execution Instructions

### Apply all pending migrations (first-time setup or after adding new files)

```bash
# Host-based (PostgreSQL running via Docker Compose on port 5433)
DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable" \
  go run ./cmd/migrate up

# Or using the pre-built binary
DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable" \
  ./bin/migrate up
```

### Check migration status

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable" \
  go run ./cmd/migrate status
```

**Safety guarantees:**
- Migrations are applied in ascending version order (001, 002, 003, 004, …)
- Each migration runs inside its own transaction; failure rolls back and halts with a non-zero exit code
- Applied versions are recorded in `schema_migrations`; already-applied migrations are never re-run
- Running `migrate up` a second time outputs `all migrations already applied — nothing to do` and exits 0

---

## Configuration Setup Instructions

### Required values (no safe default exists)

| Key | Env var | Description |
|-----|---------|-------------|
| `DATABASE_URL` | `DATABASE_URL` | PostgreSQL connection string |
| `JWT.SIGNING_KEY` | `JWT_SIGNING_KEY` | HS256 signing key — minimum 32 characters |

### Values with safe development defaults

| Key | Env var | Default |
|-----|---------|---------|
| `APP_ENV` | `APP_ENV` | `development` |
| `HTTP_ADDR` | `HTTP_ADDR` | `:8080` |
| `LOG_LEVEL` | `LOG_LEVEL` | `info` |
| `JWT.ISSUER` | `JWT_ISSUER` | `vendorsaathi` |
| `JWT.AUDIENCE` | `JWT_AUDIENCE` | `vendorsaathi-api` |
| `JWT.ACCESS_TOKEN_TTL` | `JWT_ACCESS_TOKEN_TTL` | `15m` |
| `JWT.REFRESH_TOKEN_TTL` | `JWT_REFRESH_TOKEN_TTL` | `720h` |
| `MIGRATION.PATH` | `MIGRATION_PATH` | `file://migrations` |
| `CORS_ORIGINS` | `CORS_ORIGINS` | _(empty — OK in development mode)_ |

### Host-based development (config.yaml)

```bash
# One-time setup — creates config.yaml with localhost:5433
cp config.yaml.example config.yaml
# Edit JWT.SIGNING_KEY if desired — the example value is intentionally development-only
```

### Environment-variable-only (Docker Compose / CI)

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable"
export JWT_SIGNING_KEY="your-super-secret-signing-key-change-in-production-min-32-chars"
# CORS_ORIGINS accepts a JSON array or comma-separated list
export CORS_ORIGINS='["http://localhost:3000","http://localhost:5173"]'
```

---

## Exact Commands to Start the Application

### 1. Start PostgreSQL (Docker)

```bash
docker compose up postgres -d
# Wait for healthy status
docker compose ps
```

### 2. Apply migrations

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable" \
  go run ./cmd/migrate up
```

### 3. Start the API (host-based, with config.yaml)

```bash
# Ensure config.yaml exists (copy from example if not)
cp config.yaml.example config.yaml   # already uses localhost:5433

go run ./cmd/api
# Or run the pre-built binary:
./bin/vendorsaathi
```

### 4. Start the API (env-vars only, no config.yaml)

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5433/vendorsaathi?sslmode=disable" \
JWT_SIGNING_KEY="your-super-secret-signing-key-change-in-production-min-32-chars" \
  ./bin/vendorsaathi
```

### 5. Start everything via Docker Compose

```bash
docker compose up --build
```
> The API container uses the environment variables defined in `docker-compose.yml` directly;
> no `config.yaml` is needed inside the container.

### 6. Build admin dashboard

```bash
cd admin && npm install && npm run build
```

---

## Verification Results (2026-11-06)

| Check | Result |
|-------|--------|
| `go build -o bin/vendorsaathi ./cmd/api` | ✅ Exit 0 |
| `go build -o bin/migrate ./cmd/migrate` | ✅ Exit 0 |
| `go vet ./...` | ✅ Zero issues |
| `go test ./...` | ✅ `ok vendorsaathi/tests` |
| `gofmt` on modified files | ✅ No diff |
| `docker compose config --quiet` | ✅ Valid |
| `npm run build` (admin) | ✅ 253 kB bundle |
| `./bin/migrate up` (first run) | ✅ 4 migrations applied |
| `./bin/migrate up` (second run) | ✅ Idempotent — `nothing to do` |
| `./bin/migrate status` | ✅ All 4 shown as `[applied]` |
| `GET /health/live` | ✅ `200 {"status":"ok"}` |
| `GET /health/ready` | ✅ `200 {"status":"ready"}` |
| Graceful shutdown (SIGTERM) | ✅ Server closes cleanly, DB pool closed |

---

## Phase 8: Authentication Fixes & Transactional Registration — ✅ COMPLETED (2026-11-07)

### Files Modified & Created

| File | Change |
|------|--------|
| `internal/platform/database/postgres.go` | Added `DBTX` interface and `DBFromContext(ctx, fallback)` for transaction context propagation |
| `internal/modules/auth/repository/user_repo.go` | Integrated `r.db(ctx)`, updated `CreateUser` to set status `ACTIVE`, added `CreateUserWithStatus` and `GetUserByEmail` |
| `internal/modules/customers/repository/customer_repo.go` | Integrated `r.db(ctx)` for atomic transaction participation |
| `internal/modules/auth/service/auth_service.go` | Implemented transactional registration (user + CUSTOMER role + customer profile + tokens), duplicate checks for phone/email, secure login (verifying password before checking status, requiring ACTIVE status), and safe refresh token rotation (verifying active user before revoking old session) |
| `internal/app/app.go` | Wired `customerRepo` and `txManager` into `NewAuthService` |
| `api/openapi.yaml` | Updated `/auth/register` to require `phone` & `password` with optional `email` and no `role`; added response schemas |
| `tests/auth_test.go` | **NEW** — 9 comprehensive unit tests covering registration, duplicate phone/email, profile creation, rollback, login, password check, inactive user, and refresh token lifecycle |

### Problems Fixed
1. **Simplified & Consistent Registration**:
   - Public registration accepts `phone` (required), `email` (optional), and `password` (required).
   - Removed `role` from public registration requests; public registration always assigns `CUSTOMER` role.
   - Automatically creates `customer_profiles` row in the database with default preferences.
   - Wraps user creation, role assignment, customer profile creation, and token session generation in a single database transaction via `TxManager.RunInTx`. Rollback cleans up all state if any operation fails.
   - Newly registered customer users are created with `ACTIVE` status so they can immediately authenticate and log in without SMS/OTP.
2. **Login Verification**:
   - Validates non-empty phone and password.
   - Verifies Argon2id password hash before inspecting account status (prevents user enumeration).
   - Enforces `ACTIVE` status (rejects `PENDING`, `SUSPENDED`, or `CLOSED` users with `ACCOUNT_INACTIVE`).
3. **Safe Refresh Token Rotation**:
   - Validates refresh token JWT signature and DB session validity.
   - Checks that the user is `ACTIVE` *before* revoking the existing refresh session.
   - Revokes the old session and issues a fresh access & refresh token pair.
4. **Documentation**:
   - Updated `api/openapi.yaml` with the simplified registration request, typed responses, and schemas.

---

## Remaining Blockers

None. All authentication flows, transactions, migrations, and health checks are working.

---

## Phase 9: Application Features Implementation Checklist (2026-11-08) — ✅ COMPLETED

### 1. Features Implemented & Verified

- [x] **Step 1: Codebase Inspection & Checklist**
  - Completed comprehensive codebase review of `architecture.md`, existing Go modules, Postgres schema, Flutter app, and React admin dashboard.
  - Documented feature inventory and execution plan.

- [x] **Step 2: First-Launch Language Selection & Bilingual Localization**
  - Mobile: Built `LanguageSelectionScreen` shown on first app launch ("Choose your preferred language" with "English" and "ಕನ್ನಡ").
  - Persistent storage in `FlutterSecureStorage` prior to authentication.
  - Synced to backend user profile (`users.preferred_language`) upon login/registration.
  - Settings screen allows instant switching of language at any time.
  - Admin: Built `LanguageContext` in React with full English and Kannada dictionaries, persisted in `localStorage`, with instant navbar/sidebar toggle.

- [x] **Step 3: Complete Vendor Functionality**
  - Products: Create, edit, and deactivate products with selling price, cost price (`cost_price_paise`), category, availability status, and margin calculation.
  - Inventory: Real-time stock update, low-stock alerts, stock movement history, atomic non-negative stock deduction on orders.
  - Orders: Incoming order alerts, order actions (Accept, Reject, Prepare, Ready, Out for delivery, Complete), cancellation request review.
  - Payments & Udhaar: Split payment records (Cash/UPI/Hybrid), manual UPI verification, customer credit accounts, append-only ledger entries (Credit, Repayment, Adjustment), outstanding balances.
  - Expenses: Operational expense recording and deduction for accurate net profit derivation.

- [x] **Step 4: Complete Customer Functionality**
  - Registration & Auth: Localized customer registration with phone, email, and password; seamless automatic profile creation.
  - Discovery: Spatial nearby vendor discovery with Haversine distance display, operating hours, and live catalog.
  - Shopping Experience: Category filtering, single-vendor cart enforcement with replacement dialog, quantity increment/decrement, and checkout modal.
  - Checkout & Fulfilment: Pickup vs Delivery selector, address management, loyalty points redemption, itemized price breakdown.
  - Post-Order: Order tracking timeline, cancellation request flow, star ratings & written review submission, loyalty points balance view.

- [x] **Step 5: Profit & Business Analytics**
  - Gross Profit: Calculated dynamically as `SUM((unit_price - cost_price) * quantity)` on completed order items.
  - Net Profit: Calculated as Gross Profit minus recorded operational expenses from `expenses` table.
  - Product-wise profit margins (% and absolute ₹), best-selling products, sales trends, and low-stock warnings.
  - API endpoint: `GET /api/v1/analytics/vendor/detailed`.

- [x] **Step 6: GPS & Mobility**
  - Vendor live GPS coordinate capture via Flutter `geolocator` with permission handling.
  - Nearby vendor spatial discovery using bounding-box queries and Haversine distance calculations.
  - Map-based location display and delivery route distance estimations.

- [x] **Step 7: Weather Integration**
  - OpenWeather API client with graceful fallback to Bangalore Majestic meteorological baseline.
  - Displays current temperature, weather conditions, humidity, rain probability, and actionable street-vendor planning tips.
  - API endpoint: `GET /api/v1/analytics/weather?lat=...&lng=...`.

- [x] **Step 8: AI Demand Prediction**
  - Statistical run-rate moving average engine projecting 7-day product sales demand.
  - Suggested restock quantities calculated as `max(0, predicted_demand - current_stock)`.
  - Confidence scoring explaining when trading history is insufficient (< 3 active days) without fabricating data.
  - API endpoint: `GET /api/v1/analytics/vendor/demand-forecast`.

- [x] **Step 9: Ratings, Reviews & Loyalty**
  - Customer 1–5 star ratings and written comments per completed order.
  - Schema-enforced duplicate loyalty reward prevention via `UNIQUE (account_id, order_id, transaction_type)` on `loyalty_transactions`.
  - Loyalty point earning (1 point per ₹10 spent on completed orders) and checkout redemption (1 point = ₹1 discount).
  - API endpoints: `POST /api/v1/reviews`, `GET /api/v1/reviews/vendor/:vendorId`, `GET /api/v1/reviews/admin`, `PATCH /api/v1/reviews/:id/moderate`, `GET /api/v1/loyalty/me`.

- [x] **Step 10: Admin Dashboard Completion**
  - Platform overview: Real-time metrics for total customers, total vendors, active vendors, product catalog volume, category taxonomy, order counts, gross sales, collections, and system operational health.
  - Vendor management & approval: Comprehensive vendor directory with business names, phone numbers, operating locations, product volume, and one-click Approve / Suspend / Set Pending status controls (`PATCH /api/v1/vendors/:id/status`).
  - Customer & Role management: User directory with role inspection, role assignment (`POST /api/v1/users/:id/roles`), role revocation (`DELETE /api/v1/users/:id/roles`), and account status toggling.
  - Order monitoring: Filterable order lifecycle monitoring, fulfilment mode inspection, line items, loyalty discounts, and financial settlements.
  - Review moderation: Customer star ratings and review moderation table with one-click Hide / Unhide actions (`PATCH /api/v1/reviews/:id/moderate`).
  - Internationalization: English and Kannada language toggle accessible in the top navigation/sidebar and login screen.

---

## Phase 10 — Delivery Agent Workflow, Route Generation, Live Tracking, & Razorpay Gateway

- [x] **Step 1: Database Migration 007 (`007_delivery_and_payments.up.sql`)**
  - Seeded new system role `DELIVERY_AGENT` in `roles` table.
  - Added `delivery_agent_id` foreign key column to `orders` table.
  - Created `delivery_assignments` table with lifecycle status constraint (`ASSIGNED`, `ACCEPTED`, `PICKED_UP`, `OUT_FOR_DELIVERY`, `DELIVERED`), timestamps (`assigned_at`, `accepted_at`, `picked_up_at`, `delivered_at`), and `UNIQUE (order_id)` constraint.
  - Created `delivery_tracking` table for periodic GPS coordinate updates (`latitude`, `longitude`, `updated_at`, `UNIQUE (order_id)`).

- [x] **Step 2: Delivery Agent Backend Module (`internal/modules/delivery`)**
  - Admin dispatch: `POST /api/v1/delivery/orders/:orderId/assign` assigns a registered agent to an order and updates `orders.delivery_agent_id`.
  - Admin listing: `GET /api/v1/delivery/admin/assignments` returns all assignments with joined customer, vendor, and order total details.
  - Agent assignments: `GET /api/v1/delivery/agent/assignments` returns active and completed deliveries for the authenticated agent (`RequireRole("DELIVERY_AGENT")`).
  - Status progression: `PATCH /api/v1/delivery/orders/:orderId/status` validates allowed lifecycle transitions (Assigned → Accepted → Picked Up → Out for Delivery → Delivered) and synchronizes `orders.status`.
  - Role-based authorization: Delivery agents are strictly restricted from accessing unrelated vendor/customer financial ledgers, profit analytics, and Udhaar entries.

- [x] **Step 3: Route Generation Engine (`internal/modules/mobility`)**
  - Endpoint `GET /api/v1/mobility/route` taking `origin_lat`, `origin_lng`, `dest_lat`, and `dest_lng`.
  - Google Maps Directions API integration when `MAPS.API_KEY` is configured.
  - Algorithmic fallback using Haversine calculation with a 1.3 urban winding coefficient and estimated 25 km/h transit velocity when API key is unconfigured or request fails.
  - Universal cross-platform Google Maps deep-link generation (`directions_url`) that works natively on all mobile and web browsers without API key dependencies.

- [x] **Step 4: Live Delivery Tracking Engine (`internal/modules/delivery`)**
  - Delivery agent periodic location broadcasting: `POST /api/v1/delivery/orders/:orderId/location` upserts the agent's GPS coordinates into `delivery_tracking`.
  - Client & Vendor live polling: `GET /api/v1/delivery/orders/:orderId/location` allows authorized parties (the order's customer, vendor owner, assigned agent, and admin) to retrieve the latest coordinates and update timestamp.
  - Stops automatically once delivery status reaches `DELIVERED`.

- [x] **Step 5: Online Payment Gateway Integration (Razorpay)**
  - Extended payment methods with `PaymentMethodOnline = "ONLINE"`.
  - Server-side order initiation: `POST /api/v1/payments/gateway/initiate` generates a Razorpay payment order and returns public `key_id`, `gateway_order_id`, and `amount_paise`.
  - Server-side cryptographic signature verification: `POST /api/v1/payments/gateway/verify` strictly verifies the HMAC-SHA256 signature using `RAZORPAY_KEY_SECRET`. Never trusts client-side callbacks without backend verification.
  - Duplicate payment prevention via reference ID deduplication and idempotency checks.
  - Preserves existing manual Cash, UPI, and Hybrid payment workflows intact.

- [x] **Step 6: Admin Dashboard Delivery Management (`DeliveryAgentsPage.tsx`)**
  - Dedicated "Delivery Agents" navigation tab with `Truck` icon.
  - Real-time statistics: Active Agents count, Active Deliveries in progress, and Completed Deliveries.
  - Complete assignments table showing Order ID, Customer name & address, Vendor name, Amount, Status badge, and Assigned date.
  - Live GPS tracking modal fetching current agent coordinates with direct Google Maps external link.
  - "Assign Agent to Order" modal enabling admins to dispatch any delivery order to available agents.
  - "Promote to Agent" modal enabling admins to grant the `DELIVERY_AGENT` role to registered users.
  - Full English and Kannada localization support (`navDelivery`).

- [x] **Step 7: Flutter Mobile Delivery Agent & Customer Experience**
  - `DeliveryAgentHomeScreen`:
    - Shows assigned deliveries partitioned into Active and Completed tabs.
    - Lifecycle action buttons: Accept Delivery → Pick Up Order → Out for Delivery → Mark Delivered.
    - Integrated turn-by-turn navigation via `url_launcher` Google Maps deep link.
    - `LocationUpdateService`: Background-friendly periodic GPS broadcast (every 20s) with device permission checks and graceful error recovery.
  - `CustomerHomeScreen`:
    - Checkout modal now supports Cash, UPI, and Online (Razorpay) payment method selection.
    - Online Payment Sheet handles sandbox payment initiation, secure payment verification, and instant confirmation.
    - "Track Live Delivery & ETA" bottom sheet on active orders with 10-second live polling, estimated travel time, distance in km, and real-time agent GPS coordinates.
  - Role-based routing in `main.dart` with seamless multi-role switching between Customer, Vendor, and Delivery Agent portals.

---

### Verification Summary
- **Database:** PostgreSQL schema version 7 applied cleanly with `DELIVERY_AGENT` role, `delivery_assignments`, and `delivery_tracking` tables.
- **Go Backend:** `go build -o bin/vendorsaathi ./cmd/api` compiles with 0 errors. All 21 unit tests in `tests/` pass 100%.
- **Flutter Mobile App:** `flutter analyze` reports `No issues found!`. `flutter test` passes 100%.
- **React Admin Dashboard:** `npm --prefix admin run build` compiles with 0 errors and generates optimized production bundle.
- **Live Server Test:** End-to-end HTTP API tests verified `/health/ready`, `/api/v1/mobility/route`, `/api/v1/auth/login`, `/api/v1/delivery/admin/assignments`, and `/api/v1/payments/gateway/initiate`.

---

---

## Phase 11 — Complete Migration from Google Maps to Open-Source Mapping Stack

- [x] **Step 1: Backend Architecture & Configuration (`internal/config`, `config.yaml`, `config.yaml.example`)**
  - Completely decommissioned Google Maps API dependencies and removed Google Maps API key requirements.
  - Introduced centralized, environment-driven mapping configuration:
    - `MAP_PROVIDER=maplibre`
    - `MAP_TILE_URL=https://tile.openstreetmap.org/{z}/{x}/{y}.png`
    - `ROUTING_PROVIDER=osrm`
    - `OSRM_BASE_URL=https://router.project-osrm.org`
    - `GEOCODING_PROVIDER=nominatim`
    - `GEOCODING_BASE_URL=https://nominatim.openstreetmap.org`
  - Zero private credentials required for local development and testing; seamless transition path to self-hosted OSRM and tile servers.

- [x] **Step 2: OSRM Road Routing & Nominatim Geocoding Engine (`internal/modules/mobility`)**
  - Replaced Google Directions API with OSRM (Open Source Routing Machine) `/route/v1/driving/{longitude},{latitude};{longitude},{latitude}?overview=full&geometries=geojson`.
  - Strict compliance with OSRM coordinate ordering (`{longitude},{latitude}`). Coordinate bounds validation prevents invalid parameters (-90 to +90 lat, -180 to +180 lng).
  - Returns GeoJSON LineString geometry (`coordinates: [][]float64`), road distance in meters, duration in seconds, OpenStreetMap attribution, and open-source directions URL.
  - Returns structured `routing_available: false` when OSRM cannot calculate a route without silently fabricating synthetic straight-line coordinates.
  - Nominatim integration for forward (`/mobility/geocode`) and reverse geocoding (`/mobility/reverse-geocode`) with in-memory caching and compliant `VendorSaathi/1.0` User-Agent.
  - API endpoint `/mobility/config` exposes active tile and provider metadata to frontend clients.

- [x] **Step 3: React Admin Dashboard MapLibre GL Integration**
  - Integrated `maplibre-gl` in `admin/package.json` with zero Google Maps dependencies.
  - Created `DeliveryMapModal.tsx`:
    - Renders MapLibre GL map styled with OpenStreetMap raster tiles.
    - Custom visual markers for Vendor (🏪), Customer (📍), and Delivery Agent (🛵).
    - OSRM GeoJSON route polyline layer (`osrm-route`) with casing and responsive bounds fitting (`map.fitBounds`).
    - Real-time GPS polling (15s interval) and distance / ETA indicators.
    - Prominent OpenStreetMap attribution: `© OpenStreetMap contributors`.
  - Wired into `DeliveryAgentsPage.tsx` with dedicated "Map" action button and direct link to OpenStreetMap.

- [x] **Step 4: Flutter Mobile Application MapLibre & OpenStreetMap Integration**
  - Integrated `maplibre_gl: ^0.28.0` in `mobile/pubspec.yaml`; completely removed all Google Maps SDKs and references.
  - Created reusable `OsmMapView` widget (`mobile/lib/shared/widgets/osm_map_view.dart`):
    - MapLibre GL rendering with OpenStreetMap tile source.
    - OSRM route polyline drawing, vendor/customer/agent circle markers, and nearby vendor clusters.
    - Mandatory OpenStreetMap attribution banner: `© OpenStreetMap contributors`.
    - Graceful fallback card for environments/devices with restricted graphics or rendering constraints.
    - "OSM Directions" button opening OpenStreetMap navigation via `url_launcher`.
  - Integrated into Customer and Delivery Agent experiences:
    - Customer Live Delivery Tracking sheet with real-time agent GPS coordinates, ETA card, and embedded `OsmMapView`.
    - Customer Nearby Vendors tab with interactive `OsmMapView` displaying nearby vendor locations.
    - Delivery Agent screen with "Route Preview" bottom sheet and turn-by-turn navigation link directed to OpenStreetMap.
  - Enhanced `RouteModel` and `NearbyVendorModel` in `mobile/lib/shared/models/models.dart` to parse GeoJSON coordinates and provider metadata.

- [x] **Step 5: Testing & Verification**
  - Added Go backend unit tests for OSRM coordinate ordering, bounds checking, and GeoJSON parsing (`tests/unit_test.go`).
  - Added Flutter widget & unit tests for `RouteModel`, `NearbyVendorModel`, and `OsmMapView` (`mobile/test/widget_test.dart`).
  - Clean compilation across Go (`go build`), React (`npm run build`), and Flutter (`flutter analyze`, `flutter test`).




