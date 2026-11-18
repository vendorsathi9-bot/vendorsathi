# VendorSaathi — Backend & API Audit

**Audit Date:** 2026-11-17  
**Scope:** Go/Gin backend, PostgreSQL, API contracts per architecture.md

---

## API Coverage Assessment

### Endpoint Implementation Status

| Module | Architecture Spec | Implemented | Gaps |
|--------|------------------|-------------|------|
| Auth | 4 endpoints | 4 | None |
| Users | 7 endpoints | 7 | Admin-only list/status/roles |
| Vendors | 12 endpoints | 12 | Profile creation needs VENDOR role |
| Customers | 8 endpoints | 8 | Address management complete |
| Products | 10 endpoints | 10 | Image upload auth separate |
| Inventory | 7 endpoints | 7 | Adjust/reserve/release/fulfil |
| Carts | 5 endpoints | 5 | One-vendor cart enforced |
| Orders | 11 endpoints | 11 | Full state machine |
| Payments | 7 endpoints | 7 | Split legs, verify, refunds |
| Udhaar | 6 endpoints | 6 | Append-only ledger |
| Notifications | 4 endpoints | 4 | FCM token management |
| Sync | 2 endpoints | 2 | Allowlisted operations only |
| Analytics | 7 endpoints | 7 | Vendor + admin views |
| Reviews | 4 endpoints | 4 | Create + moderation |
| Delivery | 7 endpoints | 7 | Agent + admin routes |
| Mobility | 6 endpoints | 6 | Nearby, route, tiles, geocode |
| Admin | 8 endpoints | 8 | Overview, users, vendors, orders, audit |

**Total: ~125 endpoints — All implemented per architecture.md**

---

## Authentication & Authorization Findings

### ✅ Strengths
1. **JWT + Refresh Token Rotation** — Short-lived access tokens (15min), rotating refresh tokens (30-day), stored as hashes
2. **Role-Based Access Control** — Middleware checks coarse roles, services check resource ownership
3. **Rate Limiting on Auth** — 10 req/min on register/login/refresh/logout (prevents brute force)
4. **Password Hashing** — Argon2id via `golang.org/x/crypto/argon2` with configurable params
5. **Token Claims** — Includes user_id, session_id, roles, permissions, issuer/audience validation
6. **Logout Revocation** — Refresh session revoked on logout; account suspension invalidates all

### ⚠️ Issues

| Issue | Severity | Details |
|-------|----------|---------|
| Rate limit too aggressive | High | 10 req/min blocks legitimate testing and burst usage; should be higher (e.g., 60/min) with burst allowance |
| No account lockout after failed attempts | Medium | Could allow credential stuffing; consider failed attempt tracking |
| Refresh token rotation not fully tested | Medium | Manual testing blocked by rate limits |
| No device fingerprinting on refresh | Low | Could add device_id binding for anomaly detection |

### 🔍 Code Inspection: Auth Service
- **Registration:** Creates user + CUSTOMER role + refresh session — ✅ Correct
- **Login:** Verifies password, creates new refresh session, rotates token — ✅ Correct
- **Refresh:** Validates session, rotates token, revokes old — ✅ Correct
- **Logout:** Revokes session by session_id — ✅ Correct

---

## Database Integration

### Schema Compliance

| Check | Status | Notes |
|-------|--------|-------|
| All 36 tables from architecture.md exist | ✅ | Includes audit_logs, sync_operations, etc. |
| UUID primary keys | ✅ | Using `gen_random_uuid()` |
| UTC timestamptz | ✅ | All timestamp columns use timestamptz |
| Integer paise for money | ✅ | BIGINT columns for all amounts |
| Foreign keys with RESTRICT | ✅ | On financial/historical records |
| Check constraints (money ≥ 0, qty ≥ 0, reserved ≤ on_hand) | ✅ | Present in migrations |
| Unique constraints (idempotency, phone, etc.) | ✅ | Present |
| Indexes on common filters | ✅ | products(vendor_id,active), orders(vendor_id,created_at), etc. |
| Soft delete via status | ✅ | No cascade deletes on orders/payments |

### Migration Quality

| Migration | Description | Up/Down | Status |
|-----------|-------------|---------|--------|
| 001 | Initial schema (roles, users, auth) | ✅ | Applied |
| 002 | Vendor, customer, product, category | ✅ | Applied |
| 003 | Carts, orders | ✅ | Applied |
| 004 | Payments, udhaar, notifications | ✅ | Applied |
| 005 | Business hours updated_at | ✅ | Applied |
| 006 | Features expansion (reviews, loyalty, expenses, delivery) | ✅ | Applied |
| 007 | Delivery assignments & payments enhancements | ✅ | Applied |
| 008 | Review integrity (unique constraint) | ✅ | Applied |

**All 8 migrations applied successfully. Down migrations exist for all.**

---

## Data Consistency & Transactions

### Transaction Boundaries (Per Architecture §13.1)

| Operation | Locks | Atomic Writes | Verified |
|-----------|-------|---------------|----------|
| Checkout + reservation | Cart, product, inventory rows | Order, items, inventory reserved, movements, history, idempotency | Code inspection ✅ |
| Reject order | Order + inventory rows | Status/history + release reserved + movement | Code inspection ✅ |
| Approve cancellation | Cancellation/order + inventory | Decision + CANCELLED + release once + history | Code inspection ✅ |
| Complete fulfilment | Order + inventory | Completed + decrement on-hand/reserved + movement | Code inspection ✅ |
| Payment create | Order/payment aggregate | Payment row + idempotency + audit | Code inspection ✅ |
| Refund create | Original payment/order | Refund row + refunded total + audit | Code inspection ✅ |
| Udhaar credit | Account row | Ledger entry + idempotency | Code inspection ✅ |
| Udhaar repayment | Account row | Validate balance + append repayment | Code inspection ✅ |
| Inventory adjustment | Inventory row | Projection + movement + sync result | Code inspection ✅ |

### Concurrency Protections

| Mechanism | Implemented | Notes |
|-----------|-------------|-------|
| Row-level locks (SELECT FOR UPDATE) | ✅ | In deterministic product-ID order |
| Deterministic lock ordering | ✅ | Sort product IDs before locking |
| Unique constraints as final guard | ✅ | On idempotency keys |
| Idempotency key persistence | ✅ | Actor + operation + key + request hash |
| Retry on deadlock | ⚠️ | Not observed in code; should add bounded retry |

---

## Security Observations

### ✅ Implemented
- Parameterized SQL via pgx (no string concatenation)
- CORS allowlist from config (no wildcard with credentials)
- Request body size limit (2 MiB)
- Security headers middleware (HSTS, CSP, X-Frame-Options, etc.)
- Structured logging with request ID (no sensitive data in logs)
- Rate limiting on auth endpoints
- Password hashing with Argon2id
- JWT validation (signature, issuer, audience, expiry, token type)
- Resource ownership checks at service level

### ⚠️ Concerns

| Issue | Severity | Recommendation |
|-------|----------|----------------|
| Rate limit config hardcoded (10/min) | Medium | Make configurable via env; increase for production |
| No request timeout on external calls | Medium | Add timeouts for FCM, Maps, Weather, Object Storage |
| Secrets in docker-compose.yml | Low | Use Docker secrets or env files for production |
| No audit log for auth events | Low | Log login/register/refresh failures |
| Error envelope exposes request_id | Low | Acceptable; UUID is not sensitive |

---

## API Response Structures

### Error Envelope (Per Architecture §11.2)
```json
{
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "One or more items are no longer available in the requested quantity.",
    "details": [{"field": "items[0].quantity", "reason": "available_quantity_exceeded"}],
    "request_id": "uuid"
  }
}
```
**Verified:** ✅ Implemented in `internal/platform/httpx/error.go`

### Pagination
- Uses `limit` (default 20, max 100) and opaque `cursor` — ✅ Implemented

### Idempotency
- `Idempotency-Key` header supported on checkout, payments, refunds, udhaar — ✅ Implemented

---

## Backend Defects Summary

| ID | Defect | Severity | Module | Fix |
|----|--------|----------|--------|-----|
| B-01 | Rate limit too restrictive (10/min) | High | Auth middleware | Increase to 60/min; add burst |
| B-02 | No deadlock retry logic | Medium | Database tx manager | Add bounded retry (3x) |
| B-03 | No account lockout on failed login | Medium | Auth service | Track failed attempts; lock after 5 |
| B-04 | Missing request timeouts on external adapters | Medium | FCM, Maps, Weather, Storage | Add 5-10s timeouts |
| B-05 | No integration tests for critical workflows | High | Orders, Payments, Inventory, Udhaar | Add PostgreSQL integration tests |
| B-06 | Admin endpoints lack audit logging | Low | Admin handlers | Call audit service on status changes |

---

## API Contract Compliance

| Check | Status |
|-------|--------|
| Base path `/api/v1` | ✅ |
| JSON snake_case | ✅ |
| RFC 3339 UTC timestamps | ✅ |
| REST semantics (GET/POST/PATCH/DELETE) | ✅ |
| Standard pagination | ✅ |
| Allowlisted filtering/sorting | ✅ (where implemented) |
| Stable error codes | ✅ |
| Request ID in responses | ✅ |
| Idempotency-Key support | ✅ |

---

## OpenAPI Specification

**Status: ❌ Not Found**

No `api/openapi.yaml` or generated spec found in repository. This should be generated from code or maintained manually for client validation.

---