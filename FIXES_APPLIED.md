# Fixes Applied — VendorSaathi Production Readiness Pass

**Baseline:** `f6fb93d`  
**Status:** All changes below are **uncommitted**. Nothing was pushed, deployed, or run against production data.

Each entry states what changed and why. Formatting-only edits are marked **(format)**.

---

## 1. Backend — authentication and configuration

### `internal/platform/auth/jwt.go`
- Pinned JWT verification to **HS256** using both `jwt.WithValidMethods` and an exact signing-method check. Prevents algorithm-confusion with other HMAC variants.
- Added `jwt.WithExpirationRequired()` so a token without `exp` is rejected.
- Added versioned Argon2id PHC hashing: `PasswordParams`, `DefaultPasswordParams`, `HashPasswordPHC`, `VerifyPasswordWithParams`, `verifyPHC`.
- Preserved exported `HashPassword`/`VerifyPassword` for backwards compatibility and existing tests.
- Legacy unversioned hashes still verify; callers receive `needsRehash=true` and upgrade the hash after a successful login.
- Malformed PHC strings, unsupported versions, invalid base64 and zero parameters are rejected with errors.

### `internal/modules/auth/service/auth_service.go`
- Accepts optional configured `PasswordParams` (variadic to keep existing test constructors compiling).
- Registration uses PHC hashing with configured cost.
- Login verifies PHC or legacy hashes, and transparently rehashes legacy hashes via the repository's `UpdatePassword` when available. Rehash failure does not block login.
- Invalid credentials remain a single generic error for both unknown phone and bad password.

### `internal/config/config.go`
- Added `TRUSTED_PROXIES` config/env binding with JSON-array or comma-separated parsing (shared `parseStringListEnv` helper; `CORS_ORIGINS` uses it too).
- Explicitly binds `PASSWORD_HASH_*` environment variables.
- Non-development validation:
  - `JWT_SIGNING_KEY` must be at least 32 characters.
  - Known placeholder substrings are rejected.
  - `CORS_ORIGINS` must be non-empty and must not contain `*`.
  - Refresh token TTL must exceed access token TTL.
- Argon2id parameters must be positive, with salt ≥ 8 bytes and key ≥ 16 bytes.

### `cmd/create-admin/main.go`
- Bootstrap admin hashes use `HashPasswordPHC` with the deployment's configured parameters instead of duplicated literals.

### `internal/platform/seeder/seeder.go`
- Demo hashes use `HashPasswordPHC` **(and gofmt)**. The seeder still refuses to run outside `APP_ENV=development`.

### `internal/platform/auth/password_test.go` **(new)**
- Tests PHC verify success/failure, no-rehash for current parameters, legacy verify + rehash flag, and rejection of malformed/unsupported hashes.

---

## 2. Backend — middleware and HTTP server

### `internal/app/app.go`
- `SetTrustedProxies(cfg.TrustedProxies)` with a fail-closed fallback to trusting none.
- Added `SecurityHeadersMiddleware` and a 2 MiB `BodyLimitMiddleware`.
- Added `ReadHeaderTimeout: 10s` and `MaxHeaderBytes: 1 MiB`.
- Auth routes (`/register`, `/login`, `/refresh`, `/logout`) now use a 10-requests/minute limiter.
- Added router-level `RequireRole("ADMIN", "SUPER_ADMIN")` for:
  - user list/status/role assignment/revocation,
  - vendor status updates,
  - category create/update/delete,
  - platform analytics overview.
- Wired configured password parameters into auth and user services.

### `internal/middleware/security.go` **(new)**
- Sets `X-Content-Type-Options`, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy`, `Cross-Origin-Resource-Policy`, and HSTS outside development.
- `BodyLimitMiddleware` wraps every request body in `http.MaxBytesReader`.

### `internal/middleware/cors.go`
- `Access-Control-Allow-Credentials` is sent only for an explicitly allowed, non-wildcard origin.
- Wildcard mode no longer claims credential support.
- `Vary: Origin` is always emitted when an origin is evaluated.

### `internal/middleware/rate_limit.go`
- `Retry-After` now uses `strconv.Itoa` (was an arbitrary rune).
- Safe `user_id` type assertion; keys are namespaced as `user:<id>`.

### `internal/middleware/recovery.go`
- Panic handler no longer performs an unchecked type assertion on `request_id`.
- Logs method, route, request ID and panic value; still returns a generic `500`.

---

## 3. Backend — authorization and data exposure

### `internal/modules/orders/service/order_service.go`
- Added `ensureVendorOwnsOrder`, called inside the transaction by accept, reject, prepare (generic transition), ready-for-pickup, out-for-delivery, complete and decide-cancellation flows.
- Added `GetOrderForActor`; order reads allow only an admin, the ordering customer, or the owning vendor.

### `internal/modules/orders/repository/order_repo.go`
- Added `VendorUserOwnsVendor`, `VendorUserOwnsVendorTx` and shared `vendorUserOwnsVendor` helper (supports profile ID or legacy owner user ID).
- Replaced broken `itoa` rune arithmetic with `strconv.Itoa`.
- Cursor predicate now precedes `ORDER BY` with consistent ordering.

### `internal/modules/orders/handler/order_handler.go`
- `GetOrderByID` uses `GetOrderForActor` with admin role detection.
- Generic `500` response; internal errors logged with request ID/route.

### `internal/modules/inventory/repository/inventory_repo.go`
- Added `VendorOwnsProduct` (profile ID or legacy owner user ID).
- Fixed movement cursor SQL: predicate in `WHERE`, `fmt.Sprintf` placeholders, limit applied last.

### `internal/modules/inventory/service/inventory_service.go`
- Added `ensureProductOwnership`; enforced on inventory read, movements, adjust, reserve, release and fulfil.

### `internal/modules/inventory/handler/inventory_handler.go`
- Resolves the authenticated user safely for inventory and movement reads; passes it to the service.

### `internal/modules/products/handler/product_handler.go`
- `GET /products/:id`: non-owners get `404` for inactive products and `cost_price_paise=0`.
- Category, vendor and public listing responses sanitise cost prices via `hideCostPrices`.
- Unauthenticated product browsing is forced to `active_only=true`.
- Added safe role/user extraction for cost visibility.

### `internal/modules/products/service/product_service.go`
- Added `CanViewCost` (admin, or caller resolves to the product's vendor).

### `internal/modules/products/repository/product_repo.go`
- Fixed cursor pagination: predicate in `WHERE`, row-value `(created_at, id)` cursor, consistent `ORDER BY`, `fmt.Sprintf` placeholders.

### `internal/modules/users/repository/user_repo.go`
- Rebuilt `WHERE` composition so status filters and cursor predicates combine correctly before `ORDER BY`.
- Fixed placeholders and consistent `(created_at, id)` ordering.

### `internal/modules/users/service/user_service.go`
- `ChangePassword` now uses the configured password parameters (with defaults for test compatibility) and PHC hashing.

### `internal/modules/users/handler/user_handler.go`
- No structural change; now protected additionally by router-level admin guards.

### `internal/modules/vendors/handler/vendor_handler.go`
- Generic error response instead of echoing internal errors; now protected by router-level admin guard for status changes.

---

## 4. Backend — reviews

### `internal/modules/reviews/service/review_service.go`
- Validates rating `1..5`.
- Order-linked reviews: the order must belong to the caller, match the reviewed vendor, be `COMPLETED`, and not already be reviewed.
- Order-less reviews: the customer must have a completed order with the vendor.
- Returns `403`/`409`/`422` domain errors rather than leaking internals.

### `internal/modules/reviews/repository/review_repo.go`
- Maps PostgreSQL unique violations (`23505`) to a `409` conflict.
- Added `OrderReviewEligibility`, `ReviewExistsForOrder`, `CustomerHasCompletedOrderWithVendor`.

### `internal/modules/reviews/handler/review_handler.go`
- Added `handleError` mapping app/domain errors to proper status codes; logs unexpected errors; generic `500`.

### `migrations/008_review_integrity.up.sql` **(new)**
- Unique partial index on `reviews(order_id) WHERE order_id IS NOT NULL`.
- Index for visible vendor review listings.
- Index for admin moderation listings.
- Additive/idempotent; no data rewrite or deletion.

### `migrations/008_review_integrity.down.sql` **(new)**
- Drops only the three indexes created by the up migration.

---

## 5. Backend — error handling

Generic error responses and server-side logging were applied to:

- `internal/modules/analytics/handler/analytics_handler.go` — mapping helper; no raw error text; platform overview now admin-only.
- `internal/modules/audit/handler/audit_handler.go`
- `internal/modules/notifications/handler/notification_handler.go`
- `internal/modules/sync/handler/sync_handler.go`
- `internal/modules/udhaar/handler/udhaar_handler.go`
- `internal/modules/vendors/handler/vendor_handler.go`
- Orders/reviews as described above.

---

## 6. Admin dashboard

### `admin/src/core/api/apiClient.ts`
- Stores both access and refresh tokens.
- Single-flight refresh on `401`; retries the original request once.
- Excludes login/refresh endpoints from refresh loops.
- Sends `refresh_token` on logout.
- Clears both tokens on logout, refresh failure, or unrecoverable `401`.
- `logout()` is best-effort and always clears local state.

---

## 7. Flutter mobile

### `mobile/lib/core/api/api_client.dart`
- `API_BASE_URL` from `--dart-define`; release fallback is `https://vendorsathi.onrender.com/api/v1`; debug keeps emulator/localhost defaults.
- Access and refresh tokens persisted with `flutter_secure_storage`.
- `restoreSession`, `setToken`, `setTokens`, `clearSession`, `hasSession`.
- Single refresh-and-retry on `401` for non-auth endpoints.
- Storage writes/deletes are best-effort and cannot throw into authenticated flows.
- Error bodies that are plain strings are surfaced safely.

### `mobile/lib/main.dart`
- Restores the secure session at startup, validates it with `/me`, and signs out cleanly if validation fails.
- Logout clears the secure session.

### `mobile/lib/features/auth/login_screen.dart`
- Empty credential fields by default.
- Demo quick login requires `kDebugMode` **and** explicit `ALLOW_DEMO_ACCOUNTS=true`.
- Stores both tokens after login.

### `mobile/lib/features/auth/register_screen.dart`
- Stores both access and refresh tokens after registration.

---

## 8. Runtime and environment

### `Dockerfile`
- Runtime adds an unprivileged `app` user/group.
- Binaries and migrations are copied with `--chown=app:app`.
- `USER app` before `ENTRYPOINT`.
- Added `HEALTHCHECK` against `/health/ready`.

### `.env.example`
- Rewritten as a safe template: JWT key guidance, explicit CORS, trusted proxies, Argon2id parameters, migration settings, map/routing providers, Razorpay secrets, mobile build-time URL, and admin bootstrap variables. Contains no real credentials.

---

## 9. Formatting-only changes

- `internal/modules/analytics/model/analytics.go`
- `internal/modules/payments/model/payment.go`
- `internal/modules/payments/repository/payment_repo.go`
- `internal/modules/payments/service/payment_service.go`
- `tests/unit_test.go`
- `internal/platform/seeder/seeder.go` (structured-literal alignment alongside its functional change)

These were already unformatted at the baseline. `gofmt -l .` is now clean.

---

## 10. Documentation added

- `AUDIT_REPORT.md` — findings, severity, evidence and verification results.
- `FIXES_APPLIED.md` — this file.
- `SECURITY.md` — security model, controls, residual risks and operational rules.
- `DEPLOYMENT_CHECKLIST.md` — exact production steps and smoke tests.

---

## 11. What was intentionally not changed

- API routes, request/response shapes and business workflows, except where authorization or data exposure required it.
- Existing applied migrations were not edited.
- No database reset, destructive SQL, data deletion, commit, push or deployment.
- Admin token storage was not moved out of `localStorage` (would change the auth architecture; tracked as residual risk).
- Rate limiting remains in-process (single-instance deployment assumption; tracked as residual risk).
- No dependency upgrades were performed.
