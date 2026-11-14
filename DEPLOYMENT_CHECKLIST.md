# VendorSaathi — Production Deployment Checklist

Use this checklist for every production release. It assumes the existing architecture: Go API + PostgreSQL on Render, React admin on Vercel, React landing on Vercel, Flutter mobile builds distributed separately.

**Never paste real secrets into chat, tickets, commits, screenshots or build logs.** Generate them directly in the platform's secret store or a password manager.

Production endpoints (current):

| Service | URL |
|---|---|
| API | `https://vendorsathi.onrender.com` |
| Admin | `https://vendorsathi.vercel.app` |
| Landing | `https://landing-delta-sandy-18.vercel.app` |

---

## 0. Preconditions

- [ ] You have access to the Render service, the managed PostgreSQL instance, and the Vercel projects.
- [ ] You have a database backup/snapshot from the last 24 hours.
- [ ] The release commit has passed `go test ./...`, `go vet ./...`, `go build ./...`, both npm builds, `flutter analyze` and `flutter test` (see `AUDIT_REPORT.md`).
- [ ] You have reviewed `FIXES_APPLIED.md` and `SECURITY.md`.
- [ ] No development `config.yaml` is mounted into the production service.
- [ ] You are not running `cmd/seed` against production.

---

## 1. Generate and store secrets

Run locally; copy the output directly into the Render secret field or your password manager. Do not commit it.

```bash
openssl rand -base64 48      # JWT_SIGNING_KEY
```

- [ ] `JWT_SIGNING_KEY`: at least 32 characters, random, unique to production. Must not contain `change-me`, `your-secret`, `placeholder`, `example`, etc. The API refuses to start otherwise.
- [ ] `DATABASE_URL`: production connection string. Prefer TLS (`sslmode=require` or stronger if the provider supports it). The database user should not be a superuser.
- [ ] `RAZORPAY_KEY_ID` and `RAZORPAY_KEY_SECRET`: server-side only.
- [ ] Object-storage credentials if the deployment uses external storage.
- [ ] Admin bootstrap password: generate a strong unique value and store it in the password manager before running the bootstrap CLI.

---

## 2. Configure the Render API service

Set these environment variables in the Render dashboard (or your secret manager). Values below are examples/requirements, not secrets.

### Required

```text
APP_ENV=production
HTTP_ADDR=:8080
LOG_LEVEL=info
REQUEST_TIMEOUT=30s
DATABASE_URL=<managed PostgreSQL URL>
JWT_SIGNING_KEY=<random >=32 chars>
JWT_ISSUER=vendorsaathi
JWT_AUDIENCE=vendorsaathi-api
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=720h
CORS_ORIGINS=["https://vendorsathi.vercel.app","https://landing-delta-sandy-18.vercel.app"]
```

### Important production values

```text
# Render terminates TLS and forwards client IPs. Use Render's documented
# egress/proxy CIDRs; do not guess. Correct values keep per-IP rate limiting
# working. If you intentionally run without a trusted proxy, leave this unset;
# then all clients share one ClientIP and the auth limit becomes global.
TRUSTED_PROXIES=["<render-proxy-cidr>"]

# Argon2id cost for new hashes. Defaults shown; tune only with a load test.
PASSWORD_HASH_MEMORY=65536
PASSWORD_HASH_ITERATIONS=3
PASSWORD_HASH_PARALLELISM=4
PASSWORD_HASH_SALT_LENGTH=16
PASSWORD_HASH_KEY_LENGTH=32

# Startup applies pending migrations under an advisory lock (safe/idempotent).
MIGRATION_PATH=file://migrations
MIGRATION_AUTO_APPLY=true
```

### Feature credentials

```text
RAZORPAY_KEY_ID=<...>
RAZORPAY_KEY_SECRET=<...>
MAP_PROVIDER=openfreemap
MAP_STYLE_URL=https://tiles.openfreemap.org/styles/liberty
MAP_ATTRIBUTION=© OpenFreeMap © OpenMapTiles © OpenStreetMap contributors
```

### Validation rules enforced by the API

The service **will refuse to start** in production if:

- `JWT_SIGNING_KEY` is shorter than 32 characters or looks like a placeholder.
- `CORS_ORIGINS` is empty or contains `*`.
- `JWT_REFRESH_TOKEN_TTL` is not greater than `JWT_ACCESS_TOKEN_TTL`.
- Argon2id parameters are invalid.

- [ ] All required variables are set.
- [ ] `CORS_ORIGINS` contains exact origins only.
- [ ] `TRUSTED_PROXIES` matches the actual proxy path.
- [ ] No `config.yaml` volume is mounted over the image.

---

## 3. Apply and verify migration 008

Migration `008_review_integrity` is additive: it creates three indexes and does not modify or delete data.

### Option A — automatic (recommended for this repo)

With `MIGRATION_AUTO_APPLY=true`, the API applies pending migrations during startup under a PostgreSQL advisory lock. Deploy the new image and confirm the startup log reports the migration.

### Option B — explicit CLI (recommended for a controlled window)

Run from a machine that can reach the production database, with production env vars set in your shell:

```bash
export APP_ENV=production
export DATABASE_URL='<production URL>'
export JWT_SIGNING_KEY='<the production key>'   # config validation requires it
export JWT_ISSUER=vendorsaathi
export JWT_AUDIENCE=vendorsaathi-api
export CORS_ORIGINS='["https://vendorsathi.vercel.app"]'
export MIGRATION_PATH=file://migrations

go run ./cmd/migrate status
go run ./cmd/migrate up
go run ./cmd/migrate status
```

Or use the built binary in the container:

```bash
./vendorsaathi-migrate status
./vendorsaathi-migrate up
```

### Verify (read-only SQL)

```sql
SELECT version, name, applied_at
FROM schema_migrations
ORDER BY version;

SELECT indexname
FROM pg_indexes
WHERE tablename = 'reviews'
  AND indexname IN (
    'uq_reviews_order_id',
    'idx_reviews_vendor_visible',
    'idx_reviews_hidden_created'
  );
```

Expected: version `8` recorded and all three indexes present.

### If the unique index fails

The original schema already declared `reviews.order_id` unique, so a failure is unlikely. If it occurs, inspect duplicates first:

```sql
SELECT order_id, COUNT(*)
FROM reviews
WHERE order_id IS NOT NULL
GROUP BY order_id
HAVING COUNT(*) > 1;
```

Resolve duplicates deliberately (business decision). Do **not** delete production rows automatically. No automatic cleanup is built in.

- [ ] `schema_migrations` shows version 8.
- [ ] All three indexes exist.
- [ ] No duplicate `order_id` rows remain unresolved.

---

## 4. Deploy the API

- [ ] Deploy the release commit/image to Render.
- [ ] Confirm the build uses the repository `Dockerfile` and runs as the unprivileged `app` user.
- [ ] Confirm no secrets are present in the image (`docker history`, build logs).
- [ ] Wait for the health check to pass.

Verify health:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://vendorsathi.onrender.com/health/live
curl -sS -o /dev/null -w '%{http_code}\n' https://vendorsathi.onrender.com/health/ready
```

Expected: `200` for both.

- [ ] `/health/live` returns 200.
- [ ] `/health/ready` returns 200.
- [ ] Startup logs show no configuration validation error and no migration error.
- [ ] The service is not restarting in a loop.

---

## 5. Security smoke tests

### 5.1 Security headers

```bash
curl -sS -D - -o /dev/null https://vendorsathi.onrender.com/health/ready
```

Expect:

```text
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

- [ ] Headers present.
- [ ] HSTS present (production).

### 5.2 CORS allow-list

```bash
curl -i -X OPTIONS https://vendorsathi.onrender.com/api/v1/products \
  -H 'Origin: https://vendorsathi.vercel.app' \
  -H 'Access-Control-Request-Method: GET'
```

Expect the exact origin in `Access-Control-Allow-Origin` and no wildcard.

```bash
curl -i -X OPTIONS https://vendorsathi.onrender.com/api/v1/products \
  -H 'Origin: https://evil.example' \
  -H 'Access-Control-Request-Method: GET'
```

Expect **no** `Access-Control-Allow-Origin` for the disallowed origin.

- [ ] Allowed origin reflected exactly.
- [ ] Disallowed origin not reflected.
- [ ] No `*` together with credentials.

### 5.3 Authentication failure is generic

```bash
curl -i -X POST https://vendorsathi.onrender.com/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+910000000000","password":"wrong-password"}'
```

Expect `401` with a generic `invalid credentials` message. The response must not reveal whether the phone exists or include SQL/driver text.

- [ ] Generic 401.
- [ ] No internal details.

### 5.4 Authorization boundaries

Use a valid customer/vendor account (not an admin) and verify:

```bash
# Admin-only category mutation must be 403 for a non-admin.
curl -i -X POST https://vendorsathi.onrender.com/api/v1/categories \
  -H "Authorization: Bearer <non-admin-token>" \
  -H 'Content-Type: application/json' \
  -d '{"name":"should-fail"}'

# Platform analytics must be 403 for a non-admin.
curl -i https://vendorsathi.onrender.com/api/v1/analytics/platform/overview \
  -H "Authorization: Bearer <non-admin-token>"
```

- [ ] Both return `403`.
- [ ] A second vendor cannot read or transition the first vendor's order by ID (test with two vendor accounts).
- [ ] A vendor cannot read or adjust another vendor's inventory by product ID.
- [ ] Public `GET /api/v1/products/:id` does not return a non-zero `cost_price_paise`.
- [ ] Archived products return `404` to public callers.

### 5.5 Rate limiting

Do not exhaust the production limiter during business hours. If you test it, use a controlled client and expect `429` with a numeric `Retry-After` after 10 rapid credential requests.

- [ ] `Retry-After` is a decimal number (not a control character).
- [ ] Limits are per client, not globally collapsed (requires correct `TRUSTED_PROXIES`).

---

## 6. Bootstrap the first administrator

Run `cmd/create-admin` once, ideally from a Render shell or a controlled machine with production env vars. The password is read from the environment and is never written to disk.

```bash
export APP_ENV=production
export DATABASE_URL='<production URL>'
export JWT_SIGNING_KEY='<production key>'
export JWT_ISSUER=vendorsaathi
export JWT_AUDIENCE=vendorsaathi-api
export CORS_ORIGINS='["https://vendorsathi.vercel.app"]'
export ADMIN_PHONE='+91XXXXXXXXXX'
export ADMIN_PASSWORD='<strong unique password from the password manager>'
export ADMIN_EMAIL='<admin email>'          # optional

./vendorsaathi-create-admin
```

Rules:

- [ ] Phone is E.164 (`+` followed by 8–15 digits).
- [ ] Password is strong, unique, and not one of the documented demo passwords.
- [ ] The account is created with `ACTIVE` status and `SUPER_ADMIN`.
- [ ] If the user exists, the role is ensured but the password is only changed with `ADMIN_UPDATE_PASSWORD=true` — verify this is intentional.
- [ ] Clear `ADMIN_PASSWORD` from the shell/session afterwards.
- [ ] Sign in to the admin dashboard and confirm the account works.
- [ ] Do not create shared admin accounts.

---

## 7. Deploy admin and landing

### Admin

```bash
cd admin
npm ci
npm run build
```

- [ ] `VITE_API_BASE_URL` is unset or `/api/v1` so the Vercel rewrite in `admin/vercel.json` proxies `/api/*` and `/health/*` to Render.
- [ ] If you bypass the rewrite and call Render directly, set `VITE_API_BASE_URL=https://vendorsathi.onrender.com/api/v1` **and** add the admin origin to `CORS_ORIGINS`.
- [ ] Deploy the new `dist/` to Vercel.
- [ ] Hard-refresh the admin app and verify login, token refresh (wait for access-token expiry or force a 401), and logout.
- [ ] Confirm logout revokes the refresh session server-side and clears both tokens from `localStorage`.
- [ ] Confirm no Razorpay secret or server credential appears in the bundle:
  `grep -R "RAZORPAY_KEY_SECRET\|JWT_SIGNING_KEY" dist || true`

### Landing

```bash
cd landing
npm ci
npm run build
```

- [ ] Set `VITE_SITE_URL`, `VITE_APP_DOWNLOAD_URL`, `VITE_ADMIN_DASHBOARD_URL` and `VITE_CONTACT_EMAIL` as appropriate.
- [ ] Deploy `dist/` to Vercel.
- [ ] Verify the public page loads and contains no tokens or secrets.
- [ ] The landing page does not call the API directly; it does not need CORS access.

---

## 8. Build and distribute mobile

Release builds must point at the HTTPS production API:

```bash
cd mobile
flutter pub get
flutter analyze
flutter test
flutter build appbundle --release \
  --dart-define=API_BASE_URL=https://vendorsathi.onrender.com/api/v1
```

- [ ] Release signing configured with a protected keystore (never commit it).
- [ ] `API_BASE_URL` is the production HTTPS URL.
- [ ] `ALLOW_DEMO_ACCOUNTS` is **not** set to `true` in the release pipeline.
- [ ] Demo quick-login is absent from the release UI.
- [ ] Login fields are empty by default.
- [ ] Session survives an app restart (secure storage).
- [ ] Access-token expiry is handled by refresh-and-retry.
- [ ] Logout clears the session and returns to login.
- [ ] Install the signed artifact on a clean device and test login, catalogue, cart, order and logout.
- [ ] Do not ship debug builds.

---

## 9. Post-deploy monitoring (first 24 hours)

- [ ] Watch `5xx` rate and latency in Render.
- [ ] Watch authentication failures and `429` volume. A sudden global `429` pattern usually means `TRUSTED_PROXIES` is missing or wrong.
- [ ] Check logs for `panic recovered`, `migration failed`, `invalid TRUSTED_PROXIES`, `internal server error`.
- [ ] Confirm no configuration validation warnings at startup.
- [ ] Verify database connections and migration state.
- [ ] Verify Razorpay payment initiation and signature verification in test mode before enabling live payments.
- [ ] Confirm admin and mobile can complete their primary flows.
- [ ] Check that cost prices are absent from public product responses.
- [ ] Check that order/inventory cross-tenant attempts return `403`.

---

## 10. Rollback

If the release is unhealthy:

1. Redeploy the previous API image/commit in Render.
2. Keep migration `008` applied. It is additive (indexes only) and safe for the previous application version.
3. Do **not** run the `.down.sql` automatically. Only drop the indexes after explicit review and confirmation that no current code depends on them:
   ```sql
   -- Manual, reviewed rollback only:
   DROP INDEX IF EXISTS idx_reviews_hidden_created;
   DROP INDEX IF EXISTS idx_reviews_vendor_visible;
   DROP INDEX IF EXISTS uq_reviews_order_id;
   ```
4. Revert admin/landing deployments if the regression is client-side.
5. Re-issue mobile builds only if the regression affects distributed clients.

- [ ] Previous version healthy.
- [ ] Migration state understood.
- [ ] No destructive database action taken without review.

---

## 11. Common failure modes

| Symptom | Likely cause | Action |
|---|---|---|
| API exits at startup: signing key error | Placeholder or short `JWT_SIGNING_KEY` | Generate and set a ≥32-character random key. |
| API exits: `CORS_ORIGINS must be set` / `must not contain '*'` | Missing or wildcard CORS | Set exact production origins. |
| All auth requests return `429` | `TRUSTED_PROXIES` missing/wrong; all clients share the proxy IP | Set the correct proxy CIDR and redeploy. |
| Admin requests fail CORS | Admin calls Render directly without its origin in `CORS_ORIGINS` | Add the exact admin origin, or use the Vercel rewrite. |
| Migration 008 fails on unique index | Duplicate `reviews.order_id` rows | Inspect and resolve manually; do not auto-delete. |
| Mobile release cannot reach API | Wrong/missing `API_BASE_URL` or HTTP blocked | Rebuild with the production HTTPS URL. |
| Mobile logs out on restart | Secure storage unavailable or token write failed | Check platform secure-storage support and device logs. |
| Admin session dies after 15 minutes | Refresh token missing or refresh endpoint failing | Verify both tokens are stored and `/auth/refresh` works. |
| Public product response includes cost | Old backend version still deployed | Confirm the new image is live and re-test. |
| Order/inventory cross-tenant test succeeds | Old backend version or missing service checks | Confirm deployment and run `go test`/build from the release commit. |

---

## 12. Final sign-off

- [ ] All secrets set in the platform secret store; none committed or logged.
- [ ] API deployed and healthy; config validation passed.
- [ ] Migration 008 applied and verified.
- [ ] Security headers, CORS and generic auth errors verified.
- [ ] Cross-tenant order/inventory checks verified with two accounts.
- [ ] Public cost-price/archived-product checks verified.
- [ ] Admin login/refresh/logout verified.
- [ ] Mobile release built with the production API and without demo login.
- [ ] Monitoring and rollback plan confirmed.
- [ ] Release notes record the commit, migration and manual steps performed.
