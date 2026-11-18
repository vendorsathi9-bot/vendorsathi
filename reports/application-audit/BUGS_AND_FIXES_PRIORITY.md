# VendorSaathi — Bugs & Fixes Priority List

**Audit Date:** 2026-11-17  
**Source:** Backend audit, Mobile audit, Admin audit, Functional testing

---

## Critical (P0) — Must Fix Before Release

| ID | Bug / Defect | Area | Evidence | Fix Description |
|----|--------------|------|----------|-----------------|
| **P0-1** | **Mobile app uses entirely wrong color palette** (green/mint vs brand cream/amber/brown) | Mobile App | `app_colors.dart` uses `#1B4332`, `#40916C`, `#D8F3DC` — landing page uses `#F4B400`, `#FFFCEF`, `#38200F` | Replace `AppColors` with landing palette; update `AppTheme` ColorScheme; update all component themes; update `AppDecorations` shadows |
| **P0-2** | **Zero tests on admin dashboard** — no unit, integration, e2e, or component tests | Admin Dashboard | `admin/package.json` has no test script; 0 test files found | Add Vitest + React Testing Library; add Playwright for e2e; CI integration |
| **P0-3** | **Backend test coverage: 13.6% (3/22 packages)** — critical business logic untested | Backend | Only `auth`, `migration`, `mobility` packages tested; orders, payments, inventory, udhaar untested | Add PostgreSQL integration tests for checkout, stock reservation, payments, udhaar, idempotency |
| **P0-4** | **Auth rate limit too aggressive** (10 req/min) blocks legitimate usage | Backend Auth | Manual testing blocked; returns 429 after ~10 requests | Increase to 60/min with burst allowance (e.g., 10 burst); make configurable via env |
| **P0-5** | **No OpenAPI specification** — clients cannot validate contracts | Backend API | No `api/openapi.yaml` or generated spec found | Generate from code (go-swagger/oapi-codegen) or maintain manually |

---

## High (P1) — Should Fix Before Release

| ID | Bug / Defect | Area | Evidence | Fix Description |
|----|--------------|------|----------|-----------------|
| **P1-1** | **Admin dashboard: no design system** — 80%+ inline styles, no component library | Admin Dashboard | All `.tsx` files use inline styles; `index.css` has partial classes only | Extract `src/design-system/` with tokens, components (Button, Card, Table, Input, Modal, Badge, Dropdown, Pagination) |
| **P1-2** | **Admin dashboard: not responsive** — sidebar fixed 260px, tables horizontal-scroll only | Admin Dashboard | Mobile breakpoints not handled; content overflows | Collapsible sidebar with hamburger; column-hiding tables; responsive modals |
| **P1-3** | **Admin dashboard: no global error handling** — errors only `console.log` | Admin Dashboard | No toast/error boundary; silent failures on API errors | Add `react-hot-toast` or custom provider; wrap API calls; add ErrorBoundary |
| **P1-4** | **Admin dashboard: 1.33 MB JS bundle, no code-splitting** | Admin Dashboard | Single chunk warning; no `React.lazy` | Dynamic imports per feature route with `Suspense` |
| **P1-5** | **Backend: no deadlock retry logic** in transaction manager | Backend DB | Code inspection shows no retry on `deadlock_detected` | Add bounded retry (3x) with exponential backoff in `database/tx.go` |
| **P1-6** | **Backend: no account lockout** after failed login attempts | Backend Auth | Credential stuffing possible; no failed attempt tracking | Track failed attempts in Redis/DB; lock after 5 for 15 min |
| **P1-7** | **Backend: missing request timeouts** on external adapters | Backend External | FCM, Maps, Weather, Object Storage calls can hang | Add 5-10s timeouts with context cancellation |
| **P1-8** | **Mobile: hardcoded colors** in some widgets bypass `AppColors` | Mobile App | Various widgets use `Colors.white`, `Colors.grey` directly | Audit all `Color(` usages; replace with `AppColors.*` |
| **P1-9** | **Mobile: typography misaligned** — primary font not Poppins | Mobile App | `theme.dart` uses `fontFamilyFallback` only; no Poppins in `pubspec.yaml` | Add Poppins to `pubspec.yaml` fonts; set `fontFamily: 'Poppins'` with Kannada fallback |

---

## Medium (P2) — Fix in Next Sprint

| ID | Bug / Defect | Area | Evidence | Fix Description |
|----|--------------|------|----------|-----------------|
| **P2-1** | **Admin: missing `admin-access` feature** (permission management UI) | Admin Dashboard | Architecture §10.1 lists it; not implemented | Build admin permission management page per spec |
| **P2-2** | **Admin: no CSV export** on any data table | Admin Dashboard | Architecture lists as feature; not implemented | Add export button + backend streaming endpoint |
| **P2-3** | **Admin: no configurable business hours UI** | Admin Dashboard | Architecture lists; not implemented | Build settings page for business hours |
| **P2-4** | **Admin: accessibility gaps** — icon buttons lack `aria-label`, tables lack `scope="col"`, modals lack focus trap | Admin Dashboard | Code inspection | Add ARIA labels; add `scope="col"`; implement focus trap in Modal |
| **P2-5** | **Admin: no form validation library** — manual inline validation | Admin Dashboard | Login, modals validate manually | Add `zod` + `react-hook-form`; shared schemas in `core/validation/` |
| **P2-6** | **Admin: no loading skeletons** — full-screen spinner only | Admin Dashboard | Dashboard, tables show spinner | Add skeleton loaders for KPI cards, table rows |
| **P2-7** | **Admin: empty states are text-only** — no illustrations/CTAs | Admin Dashboard | "No data" rows only | Add illustrated empty states with action buttons |
| **P2-8** | **Backend: admin endpoints lack audit logging** | Backend Admin | Status changes not logged to `audit_logs` | Call audit service on admin status updates |
| **P2-9** | **Mobile: shadow system misaligned** — generic shadows vs espresso-based | Mobile App | `app_decorations.dart` likely uses gray/black | Update all `BoxShadow` to use `AppColors.espresso` with landing alphas |
| **P2-10** | **Mobile: missing button variants** — only primary `ElevatedButton` | Mobile App | Landing has primary, secondary, ghost | Add `OutlinedButtonTheme`, `TextButtonTheme` with brand colors |
| **P2-11** | **Mobile: card elevation/depth system incomplete** | Mobile App | `CardThemeData` has `elevation: 0` only | Implement multiple card styles matching landing shadows |
| **P2-12** | **Mobile: no shared loading/empty widgets** — per-feature custom | Mobile App | Inconsistent `CircularProgressIndicator`, shimmer | Create `LoadingSkeleton`, `EmptyState` in `core/widgets/` |
| **P2-13** | **Mobile: Kannada text rendering untested** with new colors | Mobile App | Font fallback only; primary font not set | Test all screens in Kannada after palette fix |

---

## Low (P3) — Polish & Technical Debt

| ID | Bug / Defect | Area | Evidence | Fix Description |
|----|--------------|------|----------|-----------------|
| **P3-1** | **Admin: inconsistent modal sizing/spacing** | Admin Dashboard | Status change, review moderation modals vary | Single `Modal` component with `size`, `title`, `actions` props |
| **P3-2** | **Admin: token stored in memory only** (React state) | Admin Dashboard | No HttpOnly cookie; vulnerable to XSS | Move to HttpOnly cookie + CSRF token (requires backend change) |
| **P3-3** | **Backend: secrets in docker-compose.yml** | Backend Config | JWT key, DB password in plain text | Use Docker secrets or `.env` files for production |
| **P3-4** | **Backend: no audit log for auth events** | Backend Auth | Login/register/refresh failures not logged | Add audit calls in auth handler |
| **P3-5** | **Mobile: no golden tests** for visual regression | Mobile App | No screenshot testing | Add `golden_toolkit` tests for key screens |
| **P3-6** | **Mobile: tablet/landscape/foldable untested** | Mobile App | Only phone code inspection | Test and adjust layouts for 600dp+, landscape |
| **P3-7** | **Backend: rate limit config hardcoded** | Backend Config | 10/min in middleware constant | Move to config struct; load from env |

---

## Summary by Area

| Area | P0 | P1 | P2 | P3 | Total |
|------|----|----|----|----|-------|
| Mobile App | 1 | 2 | 5 | 2 | 10 |
| Admin Dashboard | 1 | 3 | 7 | 2 | 13 |
| Backend | 2 | 3 | 1 | 3 | 9 |
| **Total** | **4** | **8** | **13** | **7** | **32** |

---

## Fix Order Recommendation

1. **P0-1** (Mobile palette) — Blocks all mobile visual work
2. **P0-2** (Admin tests) — Critical for safe admin actions
3. **P0-3** (Backend tests) — Critical for financial correctness
4. **P0-4** (Rate limit) — Unblocks testing & production use
5. **P0-5** (OpenAPI) — Enables client validation
6. **P1-1** (Admin design system) — Enables all other admin UI fixes
7. **P1-5/6/7** (Backend reliability) — Concurrency, security, timeouts
8. **P1-2/3/4** (Admin responsive, errors, bundle) — Core UX
9. **P1-8/9** (Mobile hardcoded colors, typography) — Post-palette cleanup
10. Remaining P2/P3 in parallel by team capacity