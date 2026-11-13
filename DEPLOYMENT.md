# VendorSaathi — Production Deployment & Recovery Runbook

This document covers the production topology and the exact steps to deploy the
fix for the admin-login `HTTP 500` incident.

## Topology

| Component | Host | Notes |
|-----------|------|-------|
| Backend API | Render (Docker, `Dockerfile`) | `https://vendorsathi.onrender.com` |
| Database | Supabase PostgreSQL | Must be migrated before the API can serve data |
| Admin dashboard | Vercel (`admin/`) | Rewrites `/api/*` and `/health/*` to Render |

The admin dashboard talks to Render through the Vercel rewrite, so browser
requests are same-origin. The rewrite is **not** the cause of the login
failure; direct calls to Render failed the same way.

## Root cause (incident summary)

The production database was reachable but **had no application schema**
(`ERROR: relation "vendor_profiles" does not exist`, SQLSTATE `42P01`). Every
database-backed endpoint returned `HTTP 500`, including
`POST /api/v1/auth/login`.

`/health/ready` only pinged the connection, so it reported `{"status":"ready"}`
even though no tables existed. The API now:

1. Applies pending, versioned, non-destructive migrations at startup.
2. Serializes migration runs with a PostgreSQL advisory lock.
3. Reports `503` from `/health/ready` when the schema is missing or pending.
4. Returns generic error bodies and logs the internal cause server-side.

## Prerequisites

- Render service connected to this repository, Docker runtime.
- Supabase project with the production database.
- A Supabase connection string for Render. **Prefer the Session pooler or a
  direct connection** (port `5432`) over the transaction pooler (port `6543`).
  Startup migrations and the advisory lock assume session affinity.
- Admin dashboard project on Vercel with root directory `admin`.

## 1. Backend environment variables (Render)

Set these in **Render → your service → Environment**. Never commit real values.

| Variable | Required | Example / notes |
|----------|----------|-----------------|
| `APP_ENV` | yes | `production` |
| `HTTP_ADDR` | yes | `:8080` (the Dockerfile already sets it) |
| `DATABASE_URL` | yes | Supabase **session/direct** URL, include `sslmode=require` |
| `JWT_SIGNING_KEY` | yes | Random ≥32 characters, unique per environment |
| `JWT_ISSUER` | recommended | `vendorsaathi` |
| `JWT_AUDIENCE` | recommended | `vendorsaathi-api` |
| `JWT_ACCESS_TOKEN_TTL` | recommended | `15m` |
| `JWT_REFRESH_TOKEN_TTL` | recommended | `720h` |
| `CORS_ORIGINS` | yes in production | JSON array, e.g. `["https://<admin>.vercel.app"]` |
| `MIGRATION_PATH` | no | default `file://migrations` (already in the image) |
| `MIGRATION_AUTO_APPLY` | no | default `true`; leave enabled |
| `LOG_LEVEL` | no | `info` in production |

Generate a signing key locally without echoing it into shell history:

```bash
openssl rand -base64 48
```

Notes:

- `MIGRATION_AUTO_APPLY=true` is safe: migrations are versioned and recorded in
  `schema_migrations`; already-applied versions are skipped.
- If you run more than one API replica, keep `MIGRATION_AUTO_APPLY=true`; the
  advisory lock prevents concurrent application. Use a session-affinity
  connection string so the lock is held on one backend.
- Do **not** point `DATABASE_URL` at a local database in Render.

## 2. Deploy the backend

1. Commit and push the changes in this repository.
2. In Render, trigger **Manual Deploy → Deploy latest commit** (or let
   auto-deploy run).
3. Watch the deploy logs for:
   - `migration: applying ...` lines for versions `001`–`007`
   - `starting HTTP server`
4. If migrations fail, the process exits and the deploy fails. Fix the database
   connection/permissions before retrying. No destructive migration exists in
   this repository.

The Docker image now contains three binaries:

- `vendorsaathi-api` — the API (runs migrations at startup)
- `vendorsaathi-migrate` — manual `up` / `status` CLI
- `vendorsaathi-create-admin` — secure first-admin bootstrap

## 3. Verify the backend

```bash
curl -i https://vendorsathi.onrender.com/health/live
# 200 {"status":"ok"}

curl -i https://vendorsathi.onrender.com/health/ready
# 200 {"status":"ready"}  ← only after migrations are applied
# 503 {"status":"not ready"} if the schema is missing/pending

curl -i "https://vendorsathi.onrender.com/api/v1/mobility/nearby-vendors?lat=12.9767&lng=77.5713&radius_km=5"
# 200 [] (or a list), never a raw SQL error

curl -i https://vendorsathi.onrender.com/api/v1/categories
# 200 [] (or a list), never HTTP 500
```

If `/health/ready` still returns `503`, the migration step did not complete.
Check the Render logs; do not bypass the readiness check.

## 4. Create the first production administrator

There is **no** default production admin. Create one explicitly. Use a unique
password (≥12 characters) and never the documented demo password
(`AdminPassword123!`); the CLI rejects demo passwords.

### Option A — Render Shell (if available on your plan)

```bash
cd /app
ADMIN_PHONE='+91XXXXXXXXXX' \
ADMIN_PASSWORD='<unique-strong-password>' \
ADMIN_EMAIL='admin@yourdomain.com' \
./vendorsaathi-create-admin
```

### Option B — Run from a trusted machine against production

Use the same production `DATABASE_URL` and the other required config variables:

```bash
cd /path/to/repo
ADMIN_PHONE='+91XXXXXXXXXX' \
ADMIN_PASSWORD='<unique-strong-password>' \
ADMIN_EMAIL='admin@yourdomain.com' \
APP_ENV=production \
DATABASE_URL='<production-session-or-direct-url>' \
JWT_SIGNING_KEY='<same-production-signing-key>' \
CORS_ORIGINS='["https://<admin>.vercel.app"]' \
go run ./cmd/create-admin
```

The command:

- Creates the user with status `ACTIVE` and role `SUPER_ADMIN`, or
- Ensures `SUPER_ADMIN` on an existing active user (password unchanged unless
  `ADMIN_UPDATE_PASSWORD=true` is also set).
- Refuses suspended/inactive users.
- Never prints the password, hash, JWT key or database URL.

To rotate the password of an existing admin:

```bash
ADMIN_PHONE='+91XXXXXXXXXX' \
ADMIN_PASSWORD='<new-unique-strong-password>' \
ADMIN_UPDATE_PASSWORD=true \
... ./vendorsaathi-create-admin
```

## 5. Deploy and verify the admin dashboard (Vercel)

1. Vercel project root directory: `admin`.
2. Build command: `npm run build`; output directory: `dist`.
3. `admin/vercel.json` rewrites `/api/:path*` and `/health/:path*` to
   `https://vendorsathi.onrender.com`.
4. Optional `VITE_API_BASE_URL`: leave unset (defaults to `/api/v1`). Only set
   it to a full backend URL if you intentionally bypass the rewrite; then add
   the admin origin to the backend `CORS_ORIGINS`.
5. Redeploy after merging. Then open the dashboard and log in with the phone
   number and password created in step 4.

The login form no longer prefills or offers demo credentials. A 401 now shows
the server's “invalid credentials” message instead of “session expired”, and
non-admin accounts are rejected by the dashboard.

## 6. Rollback

- Application rollback: redeploy the previous Render image.
- Database: migrations `001`–`007` are additive (`CREATE TABLE`, `ADD COLUMN`,
  `CREATE INDEX`). Do **not** run `.down.sql` files in production unless you
  have taken a verified backup and intentionally want to drop data.
- Do not drop/truncate/recreate the production database to “reset” login.

## 7. Local development

```bash
docker compose up -d postgres
go run ./cmd/migrate status
go run ./cmd/api
```

The API auto-applies migrations on startup in development as well. To disable
that for a one-off run:

```bash
MIGRATION_AUTO_APPLY=false go run ./cmd/api
```

## 8. Post-deploy checklist

- [ ] Render deploy succeeded and logs show migrations applied.
- [ ] `/health/ready` returns `200 {"status":"ready"}`.
- [ ] `/api/v1/categories` returns `200`, not `500`.
- [ ] `/api/v1/mobility/nearby-vendors?lat=...&lng=...` returns `200`, not a raw SQL error.
- [ ] A `SUPER_ADMIN` exists for the intended production phone number.
- [ ] Admin login succeeds against the production API.
- [ ] `POST /api/v1/auth/login` with a wrong password returns `401`, not `500`.
- [ ] No passwords, JWT keys, database URLs or tokens appear in logs.
