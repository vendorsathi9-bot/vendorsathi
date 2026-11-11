# VendorSaathi — Public Landing Website

The public marketing site for **VendorSaathi**, a local vendor discovery, ordering,
and delivery platform. It introduces the product to customers, local vendors, and
delivery partners, and links out to the app and admin dashboard when those URLs are
configured.

This project is fully isolated inside `landing/`. It does not import from or modify
the Flutter app, Go backend, admin dashboard, or Docker setup.

## Tech stack

- React 18 + TypeScript
- Vite 5
- Tailwind CSS 3
- Framer Motion
- Lucide React
- `qrcode.react` (QR code rendered only when a real download URL exists)
- Self-hosted Poppins (Latin) and Noto Sans Kannada (Kannada subset) via `@fontsource`

## Quick start

```bash
cd landing
npm install
npm run dev
```

Vite serves the site at <http://localhost:4173>.

### Scripts

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `npm install`     | Install dependencies                     |
| `npm run dev`     | Start the local dev server on port 4173  |
| `npm run build`   | Type-check and build to `dist/`          |
| `npm run preview` | Preview the production build             |

## Environment variables

Copy `.env.example` to `.env` (or set the same variables in Vercel). Only `VITE_`
variables are exposed to the browser — never add secrets here.

| Variable                   | Required | Purpose                                                                                                 |
| -------------------------- | -------- | ------------------------------------------------------------------------------------------------------- |
| `VITE_APP_DOWNLOAD_URL`    | No       | Public Android download URL. When empty, the download section shows a **Coming soon** state, no QR code. |
| `VITE_ADMIN_DASHBOARD_URL` | No       | Public admin dashboard URL. When empty, the admin section shows an operator-access note instead of a link. |
| `VITE_CONTACT_EMAIL`       | No       | Optional public contact address. When empty, contact placeholders are shown instead of invented details. |
| `VITE_SITE_URL`            | No       | Optional canonical origin used for absolute Open Graph/canonical URLs, e.g. `https://vendorsaathi.com`. |

Invalid or non-`http(s)` URLs are ignored, so a bad value falls back to the honest
empty state rather than rendering a broken link.

## Deploying to Vercel

1. Import the repository into Vercel.
2. Set **Root Directory** to `landing`.
3. Framework preset: **Vite** (build command `npm run build`, output directory `dist`).
4. Add any of the environment variables above under **Project → Settings → Environment Variables**.
5. Deploy.

`vercel.json` already declares the build command, output directory, clean URLs, and
basic security/caching headers.

## Project structure

```
landing/
├─ index.html                 # Base SEO, Open Graph, favicon links
├─ public/                    # favicon.svg, PNG icons, og-image.png
├─ src/
│  ├─ assets/screenshots/     # Optimised WebP app screenshots (720 px wide)
│  ├─ components/
│  │  ├─ layout/              # Navbar, Footer
│  │  ├─ sections/            # Hero, HowItWorks, Features, CustomerExperience,
│  │  │                       # Vendors, DeliveryPartners, DownloadApp,
│  │  │                       # AdminDashboard, About, Contact
│  │  └─ ui/                  # Button, SectionHeading, Reveal, Logo, PhoneFrame,
│  │                          # LegalDialog
│  ├─ content/                # Site copy (site.ts) and privacy/terms (legal.ts)
│  ├─ lib/                    # cn() helper and env/config validation
│  ├─ App.tsx                 # Page composition + runtime social metadata
│  └─ index.css               # Tokens, base styles, reduced-motion handling
└─ tailwind.config.ts         # Brand palette, typography, shadows
```

## Brand and assets

The palette follows the existing VendorSaathi brand brief: golden yellow `#F4B400`,
soft yellow `#FEF4CD`, warm white `#FFFCEF`, dark espresso `#38200F`, muted brown
`#806B55`, and border `#EAE1D1`. Green is used only for semantic success states.

**Logo note:** no original VendorSaathi logo file exists in the repository yet.
The Flutter web/Android icon assets are the default Flutter launcher images. The
logo used here re-creates the storefront badge and wording from the app's splash
screen as inline SVG (`src/components/ui/Logo.tsx`, `public/favicon.svg`), using the
brand palette. Replace those two sources with the official artwork when it becomes
available.

Product screenshots are real captures from `mobile/screenshots/final/`, cropped and
compressed to WebP. Screens containing personal data or stale content were avoided.

## Content accuracy

- No invented user counts, testimonials, revenue figures, phone numbers, emails,
  social accounts, or download links.
- Feature copy describes functionality present in the mobile app and admin dashboard.
- Privacy Policy and Terms & Conditions open as accessible in-page dialogs and use
  concise, authentic policy text. They are not separate routes.
- The delivery workflow card is explicitly labelled **Illustrative**; it is not a
  screenshot.

## Accessibility and performance notes

- Single `h1`, semantic section headings, skip link, and labelled navigation.
- Mobile menu and legal dialogs support keyboard use, Escape to close, focus
  management, and body scroll locking.
- Animations respect `prefers-reduced-motion`.
- Screenshots are lazy-loaded below the fold, with explicit dimensions to avoid
  layout shift.
- Fonts are self-hosted subsets; only Latin Poppins and Kannada Noto Sans are bundled.
