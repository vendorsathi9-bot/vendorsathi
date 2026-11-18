# VendorSaathi — Mobile Application Audit (Flutter)

**Audit Date:** 2026-11-17  
**Scope:** Flutter mobile app per architecture.md MVP — Customer & Vendor experiences

---

## Build & Static Analysis

| Check | Result |
|-------|--------|
| `flutter analyze` | ✅ No issues found |
| `flutter test` (12 widget tests) | ✅ All passed |
| Code generation (freezed, json_serializable) | ✅ Configured |
| Dependencies resolved | ✅ pubspec.lock present |

---

## Architecture Compliance

### Feature Structure (Per Architecture §9.1)

| Layer | Expected | Actual | Status |
|-------|----------|--------|--------|
| `core/api` | API client, auth interceptor | ✅ `core/api/` | Complete |
| `core/auth` | Token storage, session controller | ✅ `core/auth/` | Complete |
| `core/database` | SQLite cache + migrations | ✅ `core/database/` | Complete |
| `core/errors` | App exceptions, error mapper | ✅ `core/errors/` | Complete |
| `core/localization` | i18n framework | ✅ `core/localization/` | Complete |
| `core/network` | Connectivity service | ✅ `core/network/` | Complete |
| `core/sync` | Sync runner, outbox | ✅ `core/sync/` | Complete |
| `core/widgets` | Shared widgets | ✅ `core/widgets/` | Complete |
| `features/auth` | Login, register, language | ✅ `features/auth/` | Complete |
| `features/vendor_home` | Vendor dashboard | ✅ `features/vendor_home/` | Complete |
| `features/customer_home` | Customer dashboard | ✅ `features/customer_home/` | Complete |
| `features/vendors` | Discovery, details | ✅ `features/vendors/` | Complete |
| `features/products` | Browse, details | ✅ `features/products/` | Complete |
| `features/inventory` | Stock management | ✅ `features/inventory/` | Complete |
| `features/cart` | Cart management | ✅ `features/cart/` | Complete |
| `features/orders` | Order lifecycle | ✅ `features/orders/` | Complete |
| `features/payments` | Payment recording | ✅ `features/payments/` | Complete |
| `features/udhaar` | Credit ledger | ✅ `features/udhaar/` | Complete |
| `features/notifications` | In-app notifications | ✅ `features/notifications/` | Complete |
| `features/profile` | User profile | ✅ `features/profile/` | Complete |

**All expected features implemented per architecture.**

---

## Critical Issue: Brand Color Palette Mismatch

### Landing Page Reference Palette (Source of Truth)
```dart
// From landing/tailwind.config.ts
brand: {
  yellow: '#F4B400',        // Golden amber primary
  'yellow-dark': '#DCA000',
  soft: '#FEF4CD',          // Warm cream accent
  cream: '#FFFCEF',         // Background
  white: '#FFFFFF',         // Cards
  espresso: '#38200F',      // Primary text (deep chocolate)
  brown: '#806B55',         // Secondary text (muted warm brown)
  border: '#EAE1D1',        // Soft beige borders
}
```

### Mobile App Current Palette (WRONG)
```dart
// From mobile/lib/core/design_system/app_colors.dart
AppColors {
  primary: Color(0xFF1B4332),      // Deep forest green ❌
  primaryDark: Color(0xFF133024),  // Darker green ❌
  primaryLight: Color(0xFF2D6A4F), // Lighter green ❌
  accent: Color(0xFF40916C),       // Mint/emerald ❌
  accentLight: Color(0xFF52B788),  // Bright mint ❌
  mint: Color(0xFFD8F3DC),         // Pale mint ❌
  mintSoft: Color(0xFFE8F5EE),     // Very pale mint ❌
  background: Color(0xFFFAF9F6),   // Cool off-white ❌
  surface: Color(0xFFFFFFFF),      // White ✅
  textPrimary: Color(0xFF19241E),  // Dark charcoal ❌ (should be espresso)
  textSecondary: Color(0xFF5A6B62), // Green-gray ❌ (should be warm brown)
  border: Color(0xFFE2E8E4),       // Cool gray ❌ (should be warm beige)
  // ... status colors use green/red/blue standards
}
```

### Impact
- **Entire mobile app uses a green/mint design system** instead of the brand's warm cream/amber/brown palette
- **Every screen, button, card, text, border** uses wrong colors
- **Violates explicit requirement:** "The Admin Dashboard and Flutter application UI must use the same color palette as the landing page"
- **User experience:** Jarring brand inconsistency; mobile feels like a different product

### Required Changes
1. Replace `AppColors` entirely with landing page palette
2. Update `AppTheme.lightTheme` ColorScheme to use brand colors
3. Update all component widgets using `AppColors` (buttons, cards, inputs, etc.)
4. Update `AppDecorations` shadows/borders to use espresso-based shadows
5. Verify Kannada text rendering with new colors

---

## Screen-by-Screen Functional Testing (Code Inspection)

### Authentication
| Screen | Status | Notes |
|--------|--------|-------|
| Splash | ✅ Implemented | Tests pass |
| Language Selection | ✅ Implemented | Kannada + English; tests pass |
| Login | ✅ Implemented | Phone + password; Riverpod state |
| Register | ✅ Implemented | Same as login flow |

### Customer Experience
| Screen | Status | Notes |
|--------|--------|-------|
| Home (Discovery) | ✅ Implemented | Nearby vendors, categories, search |
| Vendor List | ✅ Implemented | Grid with distance, status, ratings |
| Vendor Detail | ✅ Implemented | Products, info, hours, map link |
| Product Browse | ✅ Implemented | Category filter, search, pagination |
| Product Detail | ✅ Implemented | Image, description, price, add to cart |
| Cart | ✅ Implemented | Quantity adjust, remove, vendor switch prompt |
| Checkout | ✅ Implemented | Address select, fulfilment type, payment method |
| Order History | ✅ Implemented | List with status, pull-to-refresh |
| Order Tracking | ✅ Implemented | Status timeline, vendor contact |
| Profile | ✅ Implemented | Edit profile, addresses, language |
| Notifications | ✅ Implemented | List, mark read, FCM integration |
| Loyalty/Rewards | ✅ Implemented | Points balance, tier, history |

### Vendor Experience
| Screen | Status | Notes |
|--------|--------|-------|
| Vendor Home | ✅ Implemented | Stats, quick actions, availability toggle |
| Product Management | ✅ Implemented | CRUD, image upload, active toggle |
| Inventory Management | ✅ Implemented | Stock adjust, movements history |
| Order Management | ✅ Implemented | Accept/reject, prepare, ready, dispatch, complete |
| Sales Summary | ✅ Implemented | Charts, date range, collections vs Udhaar |
| Payments | ✅ Implemented | Record cash/UPI, verify, refunds |
| Udhaar Ledger | ✅ Implemented | Accounts, credits, repayments, balance |
| Profile | ✅ Implemented | Business info, locations, hours |

---

## Mobile-Specific Checks

| Check | Status | Notes |
|-------|--------|-------|
| Navigation & back button | ✅ | GoRouter with proper pop handling |
| State persistence | ✅ | Riverpod + SQLite cache |
| Loading indicators | ✅ | CircularProgressIndicator, shimmer skeletons |
| Empty states | ✅ | Custom empty widgets with illustrations |
| Error handling | ✅ | SnackBar + retry buttons; ErrorScreen widget |
| Form validation | ✅ | Validation on submit + real-time |
| Keyboard behavior | ✅ | Proper `textInputAction`, focus nodes |
| Network failure handling | ✅ | ConnectivityService + offline queue UI |
| App restart behavior | ✅ | Token refresh on cold start |
| Permission handling | ✅ | Location requested on-demand |
| Responsive layouts | ✅ | LayoutBuilder, MediaQuery, flexible grids |
| Secure token storage | ✅ | Flutter Secure Storage (Keychain/Keystore) |
| Offline queue UX | ✅ | Pending badge, sync status indicators |

---

## Test Coverage

### Existing Tests (12 widget tests)
| Test | Coverage |
|------|----------|
| Splash screen render | ✅ |
| Language selection | ✅ |
| Login screen | ✅ |
| RouteModel parsing | ✅ |
| NearbyVendorModel parsing | ✅ |
| OSM map attribution | ✅ |
| SalesTrendChart empty/data | ✅ |
| Kannada language switch | ✅ |
| MapStyleConfig raster/vector | ✅ |

### Missing Tests
| Area | Needed |
|------|--------|
| API integration tests | Mock API client + repository tests |
| End-to-end flows | Customer order placement, vendor order acceptance |
| Offline queue sync | Submit operation → retry → accept/reject |
| Token refresh flow | Expired access token → refresh → retry |
| Form validation edge cases | Empty, invalid, boundary values |
| Navigation guards | Role-based route protection |
| Image upload flow | Compress → upload → server response |

---

## Code Quality Observations

### ✅ Strengths
- Clean feature-based architecture with Riverpod
- Proper separation: data/domain/presentation layers
- Type-safe API client with DTOs (freezed)
- Secure storage for refresh tokens
- Offline-first with SQLite cache + outbox
- Localization framework ready (Kannada + English)
- Comprehensive error handling with user-friendly messages
- Accessibility: semantic labels, large touch targets

### ⚠️ Issues
| Issue | Severity | File |
|-------|----------|------|
| Wrong color palette (entire app) | **Critical** | `app_colors.dart`, `theme.dart`, all widgets |
| No integration test suite | High | — |
| Map UI uses OSM tiles (not Google Maps per architecture) | Medium | `mobility` feature |
| Some widgets hardcode colors instead of `AppColors` | Medium | Various |
| No golden tests for visual regression | Low | — |

---

## Device & Responsive Findings

| Device Class | Tested | Issues |
|--------------|--------|--------|
| Phone (360-400dp width) | Code inspection | Layouts use flexible grids; should work |
| Tablet (600dp+ width) | Not tested | Master-detail patterns may need adjustment |
| Foldable | Not tested | No specific handling |
| Landscape | Not tested | Some forms may overflow |

---

## Summary

**Mobile App Functional Readiness: 8/10** — All MVP features implemented, clean architecture, good offline support.

**Mobile App Brand Compliance: 1/10** — **Critical Failure**: Entire design system uses wrong color palette. Must be completely redesigned to match landing page before release.

**Recommended Priority:**
1. **Immediate:** Replace `AppColors` with landing page palette
2. **Immediate:** Update `AppTheme` ColorScheme and all component themes
3. **High:** Add integration test suite
4. **Medium:** Verify all screens with new colors; fix any hardcoded color usage
5. **Low:** Golden tests for visual regression