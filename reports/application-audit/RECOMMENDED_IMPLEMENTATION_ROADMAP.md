# VendorSaathi — Recommended Implementation Roadmap

**Audit Date:** 2026-11-17  
**Basis:** BUGS_AND_FIXES_PRIORITY.md, all audit reports  
**Goal:** Production-ready, brand-consistent, well-tested application

---

## Phase 0: Pre-Work (Week 0) — Unblock Everything

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Increase auth rate limit to 60/min + burst | Backend | 1 day | — |
| Generate OpenAPI spec from code | Backend | 2 days | — |
| Set up CI for admin dashboard (Vitest + Playwright) | Frontend | 2 days | — |
| Set up CI for mobile (flutter test + analyze) | Mobile | 1 day | — |
| Add PostgreSQL testcontainer for backend integration tests | Backend | 2 days | — |

**Exit Criteria:** Rate limit fixed; OpenAPI published; CI pipelines green for lint/build.

---

## Phase 1: Critical Brand Alignment (Week 1-2) — **Must Complete First**

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| **P0-1: Replace mobile AppColors with landing palette** | Mobile | 3 days | Phase 0 |
| Update AppTheme ColorScheme (primary, surface, background, error, etc.) | Mobile | 1 day | Palette done |
| Update all component themes (AppBar, Card, Button, Input, Divider, Chip, etc.) | Mobile | 2 days | ColorScheme done |
| Update AppDecorations shadows to espresso-based | Mobile | 1 day | Palette done |
| Audit & fix hardcoded `Color(` usages across all widgets | Mobile | 2 days | Palette done |
| Add Poppins font to pubspec.yaml; set as primary with Kannada fallback | Mobile | 1 day | — |
| Verify all screens in English + Kannada with new colors | Mobile | 2 days | All above |
| **P0-2: Admin design system extraction** | Frontend | 5 days | Phase 0 |
| Create `src/design-system/tokens.ts` with landing palette | Frontend | 1 day | — |
| Build Button, Card, Table, Input, Badge, Modal, Dropdown, Pagination components | Frontend | 3 days | Tokens done |
| Migrate all pages from inline styles to design system components | Frontend | 2 days | Components done |
| Add Poppins font to admin (npm package or local) | Frontend | 0.5 days | — |

**Exit Criteria:** Mobile app matches landing page palette exactly; admin has zero inline styles; both use Poppins + Noto Sans Kannada.

---

## Phase 2: Testing Foundation (Week 2-4) — **Parallel with Phase 1**

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| **P0-3: Backend integration tests** — checkout flow | Backend | 3 days | Phase 0 testcontainer |
| Backend integration tests — stock reservation/release | Backend | 2 days | Testcontainer |
| Backend integration tests — payments (split legs, verify, refund) | Backend | 2 days | Testcontainer |
| Backend integration tests — udhaar ledger (credit, repayment, cap) | Backend | 2 days | Testcontainer |
| Backend integration tests — idempotency keys (checkout, payments, udhaar) | Backend | 2 days | Testcontainer |
| Backend integration tests — notification outbox | Backend | 1 day | Testcontainer |
| **P0-2: Admin dashboard test suite** | Frontend | 5 days | Phase 0 CI |
| Unit tests: auth context, api client, utility functions | Frontend | 2 days | CI |
| Component tests: design system components | Frontend | 2 days | Design system done |
| E2E tests: login, vendor CRUD, order management, review moderation | Frontend | 3 days | CI |
| **P1-3: Mobile integration tests** | Mobile | 5 days | Phase 1 palette |
| API integration tests (mock client + repository) | Mobile | 2 days | — |
| E2E: customer order placement flow | Mobile | 2 days | — |
| E2E: vendor order acceptance flow | Mobile | 2 days | — |
| Offline queue sync tests | Mobile | 1 day | — |
| Token refresh flow tests | Mobile | 1 day | — |

**Exit Criteria:** Backend critical workflows ≥80% covered; admin ≥60% unit + key e2e; mobile ≥50% unit + key flows.

---

## Phase 3: Core UX & Reliability (Week 4-6)

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| **P1-5: Add deadlock retry logic** to database tx manager | Backend | 2 days | — |
| **P1-6: Implement account lockout** after 5 failed logins | Backend | 2 days | — |
| **P1-7: Add request timeouts** to FCM, Maps, Weather, Storage adapters | Backend | 2 days | — |
| **P1-2: Admin responsive sidebar** — collapsible + hamburger | Frontend | 3 days | Design system done |
| **P1-2: Admin responsive tables** — column hiding + card stacking | Frontend | 3 days | Design system done |
| **P1-3: Admin global toast/error system** | Frontend | 2 days | Design system done |
| **P1-4: Admin code-splitting** — lazy load feature routes | Frontend | 2 days | Design system done |
| **P1-8/9: Mobile hardcoded color audit + typography fix** | Mobile | 2 days | Phase 1 done |
| **P2-9: Mobile shadow system alignment** | Mobile | 1 day | Phase 1 done |
| **P2-10: Mobile button variants** (OutlinedButton, TextButton themes) | Mobile | 1 day | Phase 1 done |
| **P2-11: Mobile card depth system** (multiple elevations) | Mobile | 1 day | Phase 1 done |
| **P2-12: Mobile shared LoadingSkeleton + EmptyState widgets** | Mobile | 1 day | Phase 1 done |

**Exit Criteria:** No silent deadlocks; auth brute-force protected; external calls timeout; admin usable on mobile; errors visible to users; bundle <500KB.

---

## Phase 4: Missing MVP Features (Week 6-8)

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| **P2-1: Admin permission management** (`admin-access` feature) | Frontend + Backend | 5 days | Design system, auth |
| **P2-2: CSV export** on all admin tables + backend streaming | Frontend + Backend | 3 days | Design system |
| **P2-3: Configurable business hours UI** | Frontend + Backend | 2 days | Design system |
| **P2-4: Admin accessibility fixes** (ARIA labels, scope=col, focus trap) | Frontend | 2 days | Design system |
| **P2-5: Admin form validation library** (zod + react-hook-form) | Frontend | 2 days | Design system |
| **P2-6: Admin loading skeletons** (KPI, tables) | Frontend | 1 day | Design system |
| **P2-7: Admin illustrated empty states** with CTAs | Frontend | 2 days | Design system |
| **P2-8: Backend admin audit logging** on status changes | Backend | 1 day | — |
| **P2-13: Mobile Kannada testing** across all screens | Mobile | 2 days | Phase 1 done |

**Exit Criteria:** All architecture.md MVP features implemented; admin accessible; forms validated; empty states helpful.

---

## Phase 5: Polish & Technical Debt (Week 8-10)

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| **P3-1: Admin consistent modal component** | Frontend | 1 day | Design system |
| **P3-2: Admin token storage** — HttpOnly cookie + CSRF (requires backend) | Frontend + Backend | 3 days | — |
| **P3-3: Backend secrets** — Docker secrets / .env for production | Backend | 1 day | — |
| **P3-4: Backend auth audit logging** | Backend | 1 day | — |
| **P3-5: Mobile golden tests** for key screens | Mobile | 3 days | Phase 1 done |
| **P3-6: Mobile tablet/landscape/foldable testing** | Mobile | 2 days | Phase 1 done |
| **P3-7: Backend rate limit config** — env-driven | Backend | 0.5 days | — |
| Load testing (backend API) | Backend | 2 days | Phase 2 done |
| Security audit (penetration test) | External | 3 days | Phase 3 done |

**Exit Criteria:** Production-hardened; no plaintext secrets; visual regression guarded; mobile responsive on all targets.

---

## Phase 6: Release Preparation (Week 10-12)

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Staging deployment (full stack) | DevOps | 2 days | All phases |
| UAT with stakeholders | All | 5 days | Staging |
| Bug fix sprint | All | 3 days | UAT |
| Performance profiling + optimization | Backend + Frontend + Mobile | 3 days | Staging |
| Documentation (API, deployment, runbooks) | All | 2 days | — |
| Production deployment | DevOps | 1 day | All above |
| Post-launch monitoring (48h) | All | 2 days | Deploy |

---

## Resource Allocation (Suggested)

| Role | Phase 0 | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 | Phase 6 |
|------|---------|---------|---------|---------|---------|---------|---------|
| Backend (1-2) | 2 | 0 | 2 | 2 | 1 | 1 | 1 |
| Frontend/Admin (1-2) | 1 | 2 | 2 | 2 | 2 | 1 | 1 |
| Mobile (1-2) | 1 | 2 | 1 | 1 | 1 | 2 | 1 |
| DevOps (0.5) | 0.5 | 0 | 0.5 | 0 | 0 | 0.5 | 1 |
| QA (0.5) | 0 | 0.5 | 1 | 1 | 1 | 1 | 1 |

**Total estimated: 12-14 weeks with 4-5 engineers.**

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Mobile palette fix reveals more hardcoded colors | High | Medium | Budget 2 extra days in Phase 1 |
| Backend integration tests flaky (DB contention) | Medium | High | Use testcontainers with isolated DB per test |
| Admin design system migration breaks pages | Medium | Medium | Migrate page-by-page; visual regression tests |
| Rate limit fix insufficient for production burst | Low | High | Load test in Phase 5; adjust config |
| Kannada font rendering issues on Android/iOS | Medium | Medium | Test on real devices early in Phase 1 |
| OpenAPI spec drifts from implementation | Medium | Medium | Generate from code; CI check |

---

## Definition of Done (Per Phase)

- **Code:** Written, reviewed, merged to main
- **Tests:** Unit + integration passing in CI
- **Visual:** Mobile matches landing palette; admin uses design system
- **Docs:** Updated for any API/contract changes
- **Deploy:** Staging deployment verified

---

## Go/No-Go for Production Launch

| Criteria | Required |
|----------|----------|
| All P0/P1 bugs fixed | ✅ Yes |
| Backend integration tests ≥80% on critical paths | ✅ Yes |
| Admin dashboard tests ≥60% + key e2e | ✅ Yes |
| Mobile brand compliance 100% (palette, fonts, shadows) | ✅ Yes |
| OpenAPI spec published & validated | ✅ Yes |
| Rate limit production-ready | ✅ Yes |
| No plaintext secrets in repo | ✅ Yes |
| Load test passes (target: 500 RPS, p99 <500ms) | ✅ Yes |
| Security audit clean (no critical/high) | ✅ Yes |
| Staging UAT signed off | ✅ Yes |