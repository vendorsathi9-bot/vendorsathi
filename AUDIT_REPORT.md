# VendorSaathi — Production Readiness & Security Audit

**Audit date:** 14 November 2026  
**Repository baseline:** `f6fb93d` (all remediation is uncommitted on top of this commit)  
**Scope:** Go/Gin API + PostgreSQL migrations, React admin dashboard, React landing page, Flutter mobile app, Docker/runtime configuration, environment template.  
**Method:** Manual code review of every module, route table, repository query, migration, client auth flow and container definition; followed by formatting, build, vet, unit-test and static-analysis runs. No production data was read, modified or deleted. No migration was applied to a live database, and no deployment was performed.

---

## 1. Executive summary

The platform's core architecture, API contracts and business flows were preserved. The audit found **27 issues**: **0 Critical, 6 High, 14 Medium and 7 Low**. **21 were fully fixed, 1 was partially mitigated, and 5 remain accepted residual risks** that require an architectural or infrastructure decision (admin token storage, multi-instance rate limiting, integration-test coverage, and development-only demo assets).

The highest-impact defects were cross-tenant authorization gaps: an authenticated vendor could read another vendor's order, drive another vendor's order through its state machine, and read or mutate another vendor's inventory by guessing or observing a UUID. These are fixed with ownership checks at the service layer and actor-scoped order reads. Public catalog endpoints also returned private `cost_price_paise` values and archived products; both are now restricted.

All required verification commands pass:

| Check | Command | Result |
|---|---|---|
| Formatting | `gofmt -l .` | Clean (no output) |
| Backend build | `go build ./...` | Pass |
| Backend vet | `go vet ./...` | Pass |
| Backend tests | `go test ./...` | Pass |
| Admin production build | `npm run build` | Pass (chunk-size warning only) |
| Landing production build | `npm run build` | Pass |
| Flutter static analysis | `flutter analyze` | "No issues found!" |
| Flutter tests | `flutter test` | 12/12 passed |

Deployment is **not** blocked by code defects. It is blocked only on operator actions that cannot be performed from the repository: supplying strong production secrets, explicit CORS origins, correct trusted-proxy configuration, payment/storage credentials, and applying the additive migration. See `DEPLOYMENT_CHECKLIST.md`.

---

## 2. Severity definitions

| Severity | Meaning |
|---|---|
| Critical | Unauthenticated remote compromise, full tenant escape, or direct loss of production data/secrets. |
| High | Authenticated cross-tenant access or write, privilege escalation, credential/token compromise, or a defect that makes a release build unusable/insecure by default. |
| Medium | Information disclosure, defense-in-depth failure, availability/reliability defect, or a security control that only works in one deployment mode. |
| Low | Hardening opportunity, operational hygiene, or residual risk with limited exploitability. |

---

## 3. Issue register

### 3.1 High severity — 6 found, 6 fixed

#### H-01 — Cross-tenant order state changes (IDOR write)
- **Evidence:** `internal/modules/orders/service/order_service.go`. `AcceptOrder`, `RejectOrder`, `MarkReadyForPickup`, `MarkOutForDelivery`, `CompleteOrder` and `DecideCancellation` loaded the order by ID and transitioned it based only on the caller being an authenticated vendor user. Ownership of `order.VendorID` was never checked.
- **Impact:** Any vendor account could accept, reject, prepare, complete or decide cancellation for another vendor's orders, including triggering stock/ledger side effects.
- **Fix:** Added `VendorUserOwnsVendor` / `VendorUserOwnsVendorTx` in the order repository and `ensureVendorOwnsOrder` in the service. Every vendor-side transition now verifies ownership inside the same transaction before applying the state change. Legacy rows where `orders.vendor_id` stores the owner user ID are supported.
- **Status:** Fixed.

#### H-02 — Any authenticated user could read any order (IDOR read)
- **Evidence:** `internal/modules/orders/handler/order_handler.go`, `GetOrderByID`. The handler fetched the order without comparing the caller with the order's customer or vendor.
- **Impact:** Order contents, delivery address snapshots, phone numbers and payment state were exposed across customers and vendors.
- **Fix:** Added `GetOrderForActor` and routed the handler through it. Access is allowed only for an administrator, the customer who placed the order, or the owner of the receiving vendor.
- **Status:** Fixed.

#### H-03 — Cross-tenant inventory read and mutation (IDOR)
- **Evidence:** `internal/modules/inventory/handler/inventory_handler.go`, `internal/modules/inventory/service/inventory_service.go`. Inventory reads, movements and adjust/reserve/release/fulfil operations accepted a `productId` and used the caller's vendor ID only as a ledger actor, not as an authorization boundary.
- **Impact:** A vendor could read stock levels and movement history for, and adjust or reserve stock of, any product belonging to another vendor.
- **Fix:** Added `VendorOwnsProduct` and `ensureProductOwnership`. Every inventory read, movement listing and mutation now rejects non-owners with `403`. The handler resolves the authenticated user ID safely before calling the service.
- **Status:** Fixed.

#### H-04 — Private cost price and archived products exposed publicly
- **Evidence:** `internal/modules/products/handler/product_handler.go`. `GET /products/:id`, category listings and vendor product listings serialized the full product output, including `cost_price_paise`, and returned inactive/archived products to unauthenticated callers. `active_only=false` was honoured for public browsing.
- **Impact:** Competitors and customers could derive vendor margins, and archived/withdrawn catalogue items remained publicly discoverable.
- **Fix:** Added `ProductService.CanViewCost` and response sanitisation. Cost is returned only to an administrator or the owning vendor; public callers receive `0`. Archived products return `404` to non-owners, and public catalogue queries are forced to `active_only=true`.
- **Status:** Fixed.

#### H-05 — Release mobile app shipped demo credentials and enabled demo login by default
- **Evidence:** `mobile/lib/features/auth/login_screen.dart`. Phone and password fields were prefilled with seeded demo credentials (`+919876500001` / `VendorPassword123!`), and the demo quick-login block was enabled by `kDebugMode || ALLOW_DEMO_ACCOUNTS (default true)`.
- **Impact:** Anyone with a release build could sign in with known demo credentials against any environment where demo accounts exist, and the credentials were displayed on screen.
- **Fix:** Credential fields start empty. Quick login requires **both** `kDebugMode` and an explicit `--dart-define=ALLOW_DEMO_ACCOUNTS=true`. The seeder and `cmd/seed` already refuse to run outside `APP_ENV=development`.
- **Status:** Fixed.

#### H-06 — JWT validation and production key handling were too permissive
- **Evidence:** `internal/platform/auth/jwt.go`, `internal/config/config.go`.
  - Token parsing accepted **any** HMAC signing method (`*jwt.SigningMethodHMAC`) instead of pinning HS256.
  - Expiration was not required by the parser.
  - Non-development startup accepted short or documented placeholder signing keys, including the committed development value in `config.yaml`.
- **Impact:** Algorithm-confusion risk, non-expiring tokens if an issuer ever omitted `exp`, and production token forgery if a placeholder key was mounted or copied from the example configuration.
- **Fix:** Pinned `HS256` via both `jwt.WithValidMethods` and an exact method comparison; added `jwt.WithExpirationRequired`. Production configuration now requires a signing key of at least 32 characters and rejects known placeholder substrings (`change-me`, `your-secret`, `placeholder`, `example`, etc.). A production image built from this repository cannot start with the development key.
- **Status:** Fixed. **Manual action:** generate a real key with `openssl rand -base64 48`.

---

### 3.2 Medium severity — 14 found, 14 fixed

#### M-01 — Authentication endpoints had no rate limiting
- **Evidence:** `internal/app/app.go` registered `/auth/register`, `/auth/login`, `/auth/refresh` and `/auth/logout` without a limiter.
- **Impact:** Unbounded password guessing, credential stuffing and registration abuse.
- **Fix:** A dedicated in-memory limiter now allows 10 requests per minute per client key on all four credential endpoints. Rate-limit keys are prefixed by user ID when authenticated, and `Retry-After` is emitted correctly.
- **Status:** Fixed. Residual: per-process state (see L-06).

#### M-02 — Wildcard CORS was combined with credentials
- **Evidence:** `internal/middleware/cors.go` always sent `Access-Control-Allow-Credentials: true`, even when `Access-Control-Allow-Origin` was `*`.
- **Impact:** Invalid browser behaviour at best; reflecting credentials to arbitrary origins is unsafe if the wildcard logic changes.
- **Fix:** Credentials are sent only when the request origin exactly matches the configured allow-list. Wildcard mode sends `*` without credentials. Production configuration rejects `*` outright and requires at least one explicit origin.
- **Status:** Fixed.

#### M-03 — Missing security response headers and HSTS
- **Evidence:** No middleware set `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`, `Cross-Origin-Resource-Policy` or HSTS.
- **Impact:** MIME sniffing, framing and referrer leakage; browsers were not instructed to use HTTPS.
- **Fix:** New `internal/middleware/security.go`. Conservative headers are applied to all responses; HSTS is enabled only outside development.
- **Status:** Fixed.

#### M-04 — Unbounded request bodies, headers and header-read time
- **Evidence:** The HTTP server configured read/write/idle timeouts but no `ReadHeaderTimeout`, no `MaxHeaderBytes`, and no body limit. JSON handlers called `ShouldBindJSON` on arbitrarily large bodies.
- **Impact:** Memory exhaustion and slow-header (Slowloris-style) resource exhaustion.
- **Fix:** `ReadHeaderTimeout: 10s`, `MaxHeaderBytes: 1 MiB`, and a global 2 MiB `http.MaxBytesReader` body limit.
- **Status:** Fixed.

#### M-05 — No trusted-proxy configuration; client IP could be spoofed or collapse to one value
- **Evidence:** Gin's default trusted-proxy behaviour was left implicit. `ClientIP()` drives the auth rate limiter.
- **Impact:** If `X-Forwarded-For` is accepted from untrusted peers, an attacker can rotate spoofed IPs to bypass limits. If no proxy is trusted behind Render/nginx, every request appears to come from the proxy, turning the per-IP auth limit into a global limit.
- **Fix:** Added `TRUSTED_PROXIES` (JSON array or comma-separated CIDRs). The router calls `SetTrustedProxies` explicitly and falls back to trusting none if parsing fails.
- **Status:** Fixed. **Manual action:** set the Render/nginx egress CIDR in production.

#### M-06 — Review integrity was unenforced
- **Evidence:** `internal/modules/reviews/service/review_service.go` inserted reviews with no rating bounds, no proof that the order belonged to the caller, no check that the order was completed, no check that the order's vendor matched the reviewed vendor, and no application-level duplicate check.
- **Impact:** Fake reviews, reviews on other customers' orders, rating values outside 1–5, and trust manipulation.
- **Fix:** Added response validation, `OrderReviewEligibility`, `ReviewExistsForOrder` and `CustomerHasCompletedOrderWithVendor`. Reviews with an order must belong to the caller, match the vendor and be `COMPLETED`; order-less reviews require a prior completed order with that vendor. Duplicate inserts map PostgreSQL `23505` to `409`. Migration `008` adds a unique partial index and listing indexes without rewriting or deleting existing rows.
- **Status:** Fixed.

#### M-07 — Password hashing cost was hardcoded and hashes were not self-describing
- **Evidence:** Registration, password change, admin bootstrap and seeding each passed literal Argon2id parameters; stored hashes were an unversioned `salt || hash` blob.
- **Impact:** Raising the cost factor would invalidate every password, and there was no way to detect or upgrade weak hashes.
- **Fix:** Added versioned PHC hashes (`$argon2id$v=19$m=...,t=...,p=...$salt$hash`), configurable `PASSWORD_HASH_*` parameters, legacy verification and transparent rehash-on-successful-login. Added `password_test.go` covering PHC, legacy and malformed hashes. `cmd/create-admin`, the seeder, login, registration and password change all use the configured parameters.
- **Status:** Fixed.

#### M-08 — Admin client could not refresh sessions and did not revoke refresh tokens
- **Evidence:** `admin/src/core/api/apiClient.ts`. Any `401` cleared the access token and forced a logout; only the access token was stored; logout did not send the refresh token.
- **Impact:** Sessions died every 15 minutes, and server-side refresh sessions survived logout.
- **Fix:** Store both tokens, single-flight refresh on `401`, retry the original request once, and send `refresh_token` on logout. All tokens are cleared in a `finally` block even if the logout request fails.
- **Status:** Fixed. Residual: `localStorage` storage (see L-05).

#### M-09 — Mobile release builds targeted localhost
- **Evidence:** `mobile/lib/core/api/api_client.dart` defaulted to `http://10.0.2.2:8080` / `http://localhost:8080` for every build, with no production override except a constructor argument.
- **Impact:** A store/release build could not reach the deployed API.
- **Fix:** `API_BASE_URL` is read from `--dart-define`; release builds default to `https://vendorsathi.onrender.com/api/v1`; debug builds retain the emulator/localhost defaults.
- **Status:** Fixed.

#### M-10 — Mobile tokens were memory-only with no refresh or restore
- **Evidence:** Mobile login stored the access token in a field only; restarting the app logged the user out, and there was no refresh flow.
- **Impact:** Poor session UX and unnecessary re-authentication; no way to recover from a 15-minute access-token expiry.
- **Fix:** Tokens are persisted in `flutter_secure_storage`, restored at startup with a `/me` validation, refreshed once on `401` and retried, and cleared on logout or failed restore. Persistence failures are best-effort and cannot crash an authenticated session.
- **Status:** Fixed.

#### M-11 — Cursor pagination SQL was constructed incorrectly
- **Evidence:** `products/repository/product_repo.go`, `users/repository/user_repo.go`, `inventory/repository/inventory_repo.go`, `orders/repository/order_repo.go`. Placeholder numbers were built with `string(rune('0'+n))`, which breaks at 10+ arguments; cursor predicates were appended **after** `ORDER BY`/`LIMIT` in several queries; ordering was inconsistent with the cursor field.
- **Impact:** Paginated endpoints could return `500` or silently mis-page. Products and users also used `id`-only ordering while paginating on `(created_at, id)`.
- **Fix:** All placeholders use `fmt.Sprintf("$%d", len(args))`; predicates are added to the `WHERE` clause before ordering; ordering is consistently `(created_at DESC, id DESC)` with a matching row-value cursor; the broken `itoa` helper now uses `strconv.Itoa`.
- **Status:** Fixed.

#### M-12 — Internal error details leaked in API responses
- **Evidence:** Handlers in analytics, audit, orders, reviews, sync, notifications, udhaar and vendors returned `err.Error()` or raw driver messages with `500`.
- **Impact:** SQL fragments, constraint names and internal implementation details were disclosed to clients.
- **Fix:** Handlers now map domain errors to stable status codes and return generic messages (`internal server error`). Detailed errors are logged server-side with request ID and route. Order/analytics/review handlers log explicitly.
- **Status:** Fixed. Residual: some binding/validation errors still echo parser text, which is limited to caller-supplied input and is not sensitive.

#### M-13 — Category management and platform analytics were not admin-restricted
- **Evidence:** `internal/app/app.go`. The category route group (create/update/delete) and `GET /analytics/platform/overview` required only authentication; their services performed no role check.
- **Impact:** Any customer or vendor account could mutate the global category tree and read platform-wide business metrics.
- **Fix:** Router-level `RequireRole("ADMIN", "SUPER_ADMIN")` guards on categories, platform analytics, user administration and vendor status changes. Existing service-level checks remain as defense in depth.
- **Status:** Fixed.

#### M-14 — Container ran as root
- **Evidence:** `Dockerfile` copied binaries and migrations into `alpine` and ran as the default root user.
- **Impact:** Any process compromise inside the container had root privileges and unrestricted filesystem access.
- **Fix:** Added an unprivileged `app` user/group, `--chown=app:app` copies, `USER app`, and a `HEALTHCHECK` against `/health/ready`.
- **Status:** Fixed.

---

### 3.3 Low severity — 7 found, 1 fully fixed, 1 partially mitigated, 5 accepted

#### L-01 — Unchecked `context` type assertions in handlers
- **Evidence:** Several handlers use `c.Get("user_id").(string)` or `c.Get("roles").([]string)` after checking existence but not type.
- **Impact:** A programming error could panic. The recovery middleware converts this to `500`, so remote exploitability is low.
- **Fix:** New code uses checked assertions; the recovery middleware is panic-safe. The remaining assertions are on values written exclusively by `AuthMiddleware`, so they are internally consistent.
- **Status:** Partially fixed / accepted (defense-in-depth refactor deferred to avoid touching every handler).

#### L-02 — Panic inside the panic handler and malformed `Retry-After`
- **Evidence:** `recovery.go` asserted `request_id.(string)` inside the recovery function; `rate_limit.go` used `string(rune(retryAfter))`, producing a control character rather than a decimal value.
- **Fix:** `c.GetString("request_id")` with route/method logging; `strconv.Itoa(retryAfter)`.
- **Status:** Fixed.

#### L-03 — Development seeder creates known demo accounts
- **Evidence:** `internal/platform/seeder/seeder.go` seeds `AdminPassword123!`, `VendorPassword123!` and `CustomerPassword123!`.
- **Mitigation:** Both `cmd/seed` and `seeder.Seed` refuse to run unless `APP_ENV=development`. The production Docker image does not run the seeder. This is intentional demo data.
- **Status:** Accepted. **Manual action:** never run `cmd/seed` against production; rotate any account that was ever seeded into a shared environment.

#### L-04 — Development placeholders in committed `config.yaml`
- **Evidence:** `config.yaml` contains a development JWT key and local MinIO defaults.
- **Mitigation:** The production image copies only `config.yaml.example`, and production validation now rejects placeholder/short keys.
- **Status:** Accepted. **Manual action:** never mount the repository's `config.yaml` in production.

#### L-05 — Admin access/refresh tokens are stored in `localStorage`
- **Evidence:** `admin/src/core/api/apiClient.ts`.
- **Impact:** Any XSS in the admin origin can read both tokens. The admin app is a Vite SPA with a small dependency surface, and no XSS sink was identified during review.
- **Status:** **Residual risk, not fixed.** Moving to httpOnly, `Secure`, `SameSite` cookies requires a coordinated backend/session change and was intentionally left out of this pass to preserve the existing API contract.

#### L-06 — Rate limiter state is per process
- **Evidence:** `internal/middleware/rate_limit.go` uses an in-memory map.
- **Impact:** With multiple API instances, the effective limit is multiplied by the instance count; restarts reset counters.
- **Status:** **Residual risk, not fixed.** Acceptable for the current single-instance deployment. Use a shared Redis-backed limiter before horizontal scaling.

#### L-07 — No integration/E2E tests for authorization boundaries
- **Evidence:** The repository has unit tests for auth/password/JWT and mobile widgets, but no database-backed HTTP tests for the IDOR fixes.
- **Status:** **Residual gap, not fixed.** The added ownership checks are small and compile-tested, but a test database harness is needed to prevent regressions. Recommended next step.

---

## 4. Positive findings (no change required)

- Razorpay payment confirmation verifies an HMAC-SHA256 signature with the server-side secret before recording a payment; the secret key is never sent to clients.
- Cart and checkout operations are scoped to the authenticated customer at the service layer.
- Udhaar and delivery flows already had role/ownership checks appropriate to their use cases.
- Migrations are versioned, applied under an advisory lock, and the new migration is additive only.
- Refresh tokens are stored as hashes server-side and sessions are revocable.
- The API does not log request bodies or credentials.
- Admin and landing production builds pass. A separate dependency vulnerability scan (`npm audit`) was not part of this pass; run it as part of the release pipeline.

---

## 5. Database and migration safety

- **No existing migration was edited.** All schema changes are in new files:
  - `migrations/008_review_integrity.up.sql`
  - `migrations/008_review_integrity.down.sql`
- Migration `008` is additive and idempotent (`IF NOT EXISTS`): one unique partial index on `reviews(order_id)` and two listing indexes. It does **not** update or delete rows.
- No database was reset, truncated or modified during this audit. The seeder's destructive reset path still requires `CONFIRM_DB_RESET=yes` and `APP_ENV=development`.
- `MIGRATION_AUTO_APPLY=true` will apply `008` at API start. The migration is safe to run while the old version is serving traffic. Verify with `cmd/migrate` or the startup logs.
- If existing duplicate reviews for the same `order_id` exist, the unique index creation would fail. The original schema already declared `order_id` unique, so this is unlikely; if it occurs, resolve duplicates manually before applying. No automatic cleanup was added by design.

---

## 6. Verification log

Executed from the repository root unless noted:

```
gofmt -l .                     -> no output
go build ./...                 -> success
go vet ./...                   -> success
go test ./...                  -> all packages pass
```

Client builds:

```
cd admin   && npm run build    -> success (vite chunk-size warning only)
cd landing && npm run build    -> success
cd mobile  && flutter analyze  -> No issues found!
cd mobile  && flutter test     -> 12/12 passed
```

Known non-blocking warnings:

- Admin bundle is ~1.33 MB minified (~365 kB gzip) and Vite warns about chunks over 500 kB. This is a performance follow-up (route-level code splitting), not a correctness or security failure.
- `gofmt` was applied to five pre-existing files (`analytics/model`, `payments/model`, `payments/repository`, `payments/service`, `tests/unit_test.go`) so the repository is now format-clean. Those edits are whitespace/alignment only.

---

## 7. Remaining blockers and residual risk

| Item | Type | Owner action |
|---|---|---|
| Strong `JWT_SIGNING_KEY` (≥32 chars, no placeholder) | Deployment blocker | Set in the platform secret store. The API refuses to start in production without it. |
| Explicit `CORS_ORIGINS` | Deployment blocker | Set the exact admin/landing/mobile-web origins; `*` is rejected. |
| `TRUSTED_PROXIES` | Deployment prerequisite | Set the Render/nginx CIDR, otherwise auth rate limiting is global and client IPs are wrong. |
| Production database + migration `008` | Deployment prerequisite | Apply via startup auto-migrate or `cmd/migrate`; verify the unique index. |
| Razorpay and storage credentials | Feature prerequisite | Set server-side env vars; never ship the secret to clients. |
| Admin tokens in `localStorage` (L-05) | Accepted residual | Consider httpOnly cookie sessions in a future architectural change. |
| In-memory rate limiter (L-06) | Accepted residual | Move to Redis before running more than one API replica. |
| No DB-backed authz integration tests (L-07) | Accepted gap | Add an ephemeral PostgreSQL test harness. |
| Demo seed accounts (L-03) and `config.yaml` placeholders (L-04) | Operational | Never seed or mount development config in production. |

No source-level defect remains that prevents a production deployment once the manual items above are completed.

---

## 8. Files changed

See `FIXES_APPLIED.md` for the complete per-file rationale. Security and deployment guidance is in `SECURITY.md` and `DEPLOYMENT_CHECKLIST.md`. All changes are uncommitted relative to `f6fb93d`.
