# VendorSaathi — Security Guide

This document describes the security controls in the current codebase, the operating rules that must be followed in production, and the known residual risks. It is written for maintainers and operators, not end users.

---

## 1. Security model at a glance

| Layer | Control |
|---|---|
| Transport | HTTPS terminated by the hosting platform; HSTS sent by the API outside development. |
| Authentication | Phone + password; Argon2id password hashes; HS256 JWT access tokens; revocable refresh tokens. |
| Authorization | Bearer token → role/permission claims → route guards + service-level ownership checks. |
| Tenant isolation | Vendor ownership is checked for orders and inventory at the service layer, inside the same transaction for state changes. |
| Data exposure | Cost prices and archived products are hidden from public callers; internal errors are not returned to clients. |
| Abuse prevention | Credential endpoints limited to 10 requests/minute/client; global 2 MiB body limit; header limits and timeouts. |
| Secrets | Environment variables only; no credentials in source, images or client bundles. |
| Database | Versioned, additive migrations; advisory-locked, idempotent runner; no destructive operations in this pass. |
| Clients | Mobile tokens in OS secure storage; admin tokens in browser storage (see residual risks). |

---

## 2. Secrets management

### Required production secrets

| Variable | Requirement |
|---|---|
| `JWT_SIGNING_KEY` | Random, **≥32 characters**, unique per environment. Generate with `openssl rand -base64 48`. Must not contain `change-me`, `your-secret`, `placeholder` or `example`. |
| `DATABASE_URL` | Managed PostgreSQL URL with TLS (`sslmode=require` or stronger) where the provider supports it. |
| `CORS_ORIGINS` | Explicit, exact origins. No `*`. |
| `TRUSTED_PROXIES` | CIDR of the reverse proxy/load balancer that sets `X-Forwarded-For`. |
| `RAZORPAY_KEY_ID` / `RAZORPAY_KEY_SECRET` | Server-side only. The secret must never reach admin, landing or mobile bundles. |
| Object-storage credentials | Server-side only; not used by public clients. |

### Rules

- Never commit real secrets. `.env.example` is a template and contains no credentials.
- Never bake secrets into Docker images or Flutter `--dart-define` values. Anything compiled into a client is public.
- Store secrets in the platform's secret manager (Render environment group, Vault, SSM, etc.), not in a repository file.
- The repository's `config.yaml` is development-only. The production image copies `config.yaml.example`, and production startup rejects placeholder JWT keys. Do not mount the repository's `config.yaml` in production.
- Rotate `JWT_SIGNING_KEY` if it is ever exposed. Rotation invalidates all access tokens; refresh sessions remain valid until revoked, so also revoke sessions if compromise is suspected.
- Rotate database and Razorpay credentials immediately if a leak is suspected, then redeploy.

---

## 3. Authentication

### Passwords

- Hashed with **Argon2id**.
- New hashes use the versioned PHC format: `$argon2id$v=19$m=...,t=...,p=...$salt$hash`.
- Cost parameters come from `PASSWORD_HASH_*` (defaults: 64 MiB, 3 iterations, parallelism 4, 16-byte salt, 32-byte key).
- Legacy unversioned hashes still verify and are transparently rehashed after a successful login.
- Verification uses constant-time comparison.
- Failed login returns one generic `invalid credentials` error for unknown phone and wrong password.
- Password change verifies the current password first.
- The CLI admin bootstrap reads the password from `ADMIN_PASSWORD`, refuses documented demo passwords, and never writes it to disk.

### Access tokens

- HS256 only; algorithm is pinned in the verifier.
- `exp`, `iss`, `aud` are validated; expiration is mandatory.
- Default TTL: 15 minutes (`JWT_ACCESS_TOKEN_TTL`).
- Claims include user ID, session ID, roles and permissions.

### Refresh tokens

- Opaque JWT refresh token with a longer TTL (default 720h, must exceed access TTL).
- The server stores only a **SHA-256 hash** of the refresh token in `refresh_sessions`.
- Refresh rotates the token: the old session is revoked and a new pair is issued.
- Logout revokes the refresh session by hash.
- Revoked/expired sessions are rejected.

### Rate limiting

- `/auth/register`, `/auth/login`, `/auth/refresh` and `/auth/logout` allow 10 requests/minute per client key.
- The key is the authenticated user ID when present, otherwise `ClientIP()`.
- `Retry-After` is returned as a decimal number of seconds.
- The limiter is in-process; see residual risks.

---

## 4. Authorization and tenant isolation

### Roles

`ADMIN` and `SUPER_ADMIN` are required for:
- user administration (list, status, role assignment/revocation),
- vendor status changes,
- category management,
- platform analytics.

`VENDOR`, `CUSTOMER` and `DELIVERY_AGENT` capabilities are enforced by the relevant services.

### Ownership checks

- **Orders:** a vendor-side transition requires ownership of `order.VendorID`. The check runs inside the same transaction as the state change. Order reads allow only an admin, the ordering customer, or the owning vendor.
- **Inventory:** reads, movement history and all mutations require ownership of the product's vendor.
- **Products:** cost price is visible only to an admin or the owning vendor; archived products are hidden from non-owners.
- **Reviews:** an order-linked review requires the caller's own order, a matching vendor, `COMPLETED` status and no prior review. Order-less reviews require a prior completed order with that vendor.
- **Carts, udhaar, delivery, payments:** existing service-level checks were reviewed and retained.

Legacy data compatibility: some rows store the owner user ID directly in `vendor_id`. Ownership helpers accept both the canonical vendor-profile ID and the legacy owner-user ID.

### HTTP status behaviour

- `401` unauthenticated / invalid token.
- `403` authenticated but not permitted.
- `404` resource hidden from the caller (for example, archived product).
- `409` state/duplicate conflict.
- `422` validation failure.
- `500` generic `internal server error`; details go to server logs only.

---

## 5. Transport, CORS and browser hardening

### CORS

- Allowed origins are exact-match values from `CORS_ORIGINS`.
- Credentials are sent only for an explicitly allowed origin.
- Wildcard mode never claims credential support.
- Production rejects `*` during configuration validation.
- Set the admin, landing and any browser-based client origins explicitly.

### Security headers

The API sets:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
Permissions-Policy: geolocation=(self), camera=(), microphone=()
Cross-Origin-Resource-Policy: cross-origin
Strict-Transport-Security: max-age=31536000; includeSubDomains   # non-development only
```

HSTS assumes HTTPS termination in front of the API. Do not enable it on a plain-HTTP internal endpoint.

### Request limits

- `ReadHeaderTimeout`: 10s.
- `MaxHeaderBytes`: 1 MiB.
- Request body: 2 MiB (`http.MaxBytesReader`).

### Proxy handling

- With no `TRUSTED_PROXIES`, Gin ignores `X-Forwarded-For` and uses the socket peer address.
- Behind Render/nginx, set `TRUSTED_PROXIES` to the proxy CIDR. If omitted, every request appears to originate from the proxy, so the per-IP auth limit becomes a single global bucket.

---

## 6. Data protection and logging

- Passwords, refresh-token hashes and JWT signing keys are never logged.
- Order and inventory responses are scoped to the caller.
- Cost prices are private commercial data and are hidden from public responses.
- Error responses do not include SQL, constraint names or driver text.
- Server logs include request ID, route and method for triage; request bodies are not logged.
- Database connections should use TLS in production.
- Backups and snapshots must be encrypted and access-controlled by the database provider.
- Do not place production dumps, `.env` files or secrets in the repository or CI artifacts.

---

## 7. Mobile and admin client security

### Flutter mobile

- Access and refresh tokens are stored in `flutter_secure_storage` (Keychain/Keystore-backed where available).
- Tokens are restored at startup and validated with `/me`; a failed validation clears the session.
- A single refresh-and-retry occurs on `401` for non-auth endpoints.
- Logout clears the secure session.
- Release builds use HTTPS. Debug builds may use emulator HTTP; do not ship a debug build.
- Demo quick login requires **both** `kDebugMode` and `--dart-define=ALLOW_DEMO_ACCOUNTS=true`. Never enable it in a release pipeline.
- Build release artifacts with an explicit API URL:
  `flutter build apk --release --dart-define=API_BASE_URL=https://vendorsathi.onrender.com/api/v1`

### React admin

- Access and refresh tokens are stored in `localStorage` and sent as a Bearer token.
- A single-flight refresh handles `401`; logout sends the refresh token and clears local state.
- `localStorage` is readable by any script on the origin, so XSS is equivalent to token theft. Keep the dependency surface small, avoid `dangerouslySetInnerHTML`, and consider an httpOnly-cookie session in a future architectural change.
- Never embed the Razorpay secret or any server credential in the admin bundle.

### Landing page

- The landing page is public and must not receive or store tokens or secrets.
- Keep it on a separate origin and include that origin in `CORS_ORIGINS` only if it directly calls the API.

---

## 8. Database and migration security

- Migrations are versioned files; applied versions are recorded in `schema_migrations`.
- The runner takes a PostgreSQL advisory lock, so concurrent replicas cannot apply the same migration twice.
- Applied migrations must not be edited. Add a new numbered migration instead.
- Migration `008` is additive and idempotent; it creates indexes only.
- No destructive SQL, resets, truncations or data deletion are part of this pass.
- The development seeder refuses to run outside `APP_ENV=development`; the destructive reset path also requires `CONFIRM_DB_RESET=yes`.
- Production database users should have only the privileges the application needs. Avoid using a superuser for the API.
- Verify that migrations have run before sending traffic to a new release.

---

## 9. Deployment security checklist (summary)

The full procedure is in `DEPLOYMENT_CHECKLIST.md`. Minimum requirements:

1. Strong, unique `JWT_SIGNING_KEY` (≥32 chars).
2. Explicit `CORS_ORIGINS`; no `*`.
3. Correct `TRUSTED_PROXIES`.
4. TLS-terminating proxy in front of the API.
5. Database credentials with least privilege and TLS.
6. Razorpay/storage secrets supplied server-side only.
7. Migration `008` applied and verified.
8. No development `config.yaml`, seeder run, or demo accounts in production.
9. Admin and mobile release builds pointed at the production HTTPS API.
10. Logs and alerts monitored for `5xx`, auth failures and rate-limit spikes.

---

## 10. Residual risks (accepted in this pass)

| Risk | Why accepted | Recommended follow-up |
|---|---|---|
| Admin tokens in `localStorage` | Moving to httpOnly cookies changes the API/session contract. | Add a cookie-based admin session with CSRF protection. |
| In-memory rate limiter | Current deployment is a single API instance. | Use Redis or the platform's edge rate limiting before scaling horizontally. |
| No DB-backed authorization integration tests | Requires an ephemeral PostgreSQL harness. | Add tests for cross-tenant order/inventory/review access. |
| Development demo accounts and `config.yaml` placeholders remain in the repo | Needed for local development; both are guarded or excluded from production. | Keep environments separated; never seed or mount config in production. |
| Some validation errors echo parser text | Text is derived from caller input and is not sensitive. | Normalize messages if a stricter contract is desired. |
| Unchecked context assertions in older handlers | Values are written by trusted middleware; recovery prevents crashes. | Refactor to shared typed context helpers. |

---

## 11. Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Report it privately to the maintainers with:
- affected endpoint or component,
- reproduction steps,
- impact assessment,
- any suggested fix.

Rotate affected secrets immediately if credentials may have been exposed.
