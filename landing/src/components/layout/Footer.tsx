import { ArrowUpRight, Mail } from 'lucide-react';

import type { LegalDocumentId } from '../../content/legal';
import {
  hasAdminDashboard,
  hasContactEmail,
  siteConfig,
} from '../../lib/config';
import { Logo } from '../ui/Logo';

interface FooterProps {
  onOpenLegal: (id: LegalDocumentId) => void;
}

interface FooterLink {
  label: string;
  href: string;
  external?: boolean;
}

const platformLinks: FooterLink[] = [
  { label: 'Home', href: '#home' },
  { label: 'How It Works', href: '#how-it-works' },
  { label: 'Features', href: '#features' },
  { label: 'About', href: '#about' },
  { label: 'Contact', href: '#contact' },
];

const customerLinks: FooterLink[] = [
  { label: 'Customer experience', href: '#customers' },
  { label: 'Download the app', href: '#download' },
  { label: 'Delivery partners', href: '#delivery' },
];

const vendorLinks: FooterLink[] = [
  { label: 'For local vendors', href: '#vendors' },
  { label: 'Admin overview', href: '#admin' },
  ...(hasAdminDashboard
    ? [
        {
          label: 'Admin dashboard',
          href: siteConfig.adminDashboardUrl,
          external: true,
        },
      ]
    : []),
];

function FooterLinkList({ links }: { links: FooterLink[] }) {
  return (
    <ul className="mt-4 space-y-2.5">
      {links.map((link) => (
        <li key={link.label}>
          <a
            href={link.href}
            {...(link.external
              ? { target: '_blank', rel: 'noopener noreferrer' }
              : {})}
            className="group inline-flex items-center gap-1.5 text-sm text-brand-cream/65 transition-colors hover:text-brand-yellow"
          >
            {link.label}
            {link.external ? (
              <ArrowUpRight
                className="h-3.5 w-3.5 opacity-60"
                aria-hidden="true"
              />
            ) : null}
          </a>
        </li>
      ))}
    </ul>
  );
}

export function Footer({ onOpenLegal }: FooterProps) {
  const year = new Date().getFullYear();

  return (
    <footer className="relative overflow-hidden bg-brand-espresso text-brand-cream">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-24 top-0 h-72 w-72 rounded-full bg-brand-yellow/10 blur-3xl"
      />

      <div className="container-page relative py-14 sm:py-16">
        <div className="grid gap-10 lg:grid-cols-[1.6fr_1fr_1fr_1.3fr] lg:gap-12">
          <div>
            <a
              href="#home"
              aria-label="VendorSaathi — back to top"
              className="inline-flex rounded-2xl"
            >
              <Logo tone="light" />
            </a>
            <p className="mt-5 max-w-sm text-sm leading-relaxed text-brand-cream/65">
              A local commerce platform connecting neighbourhood vendors with
              nearby customers — with ordering, payments, vendor tools, and
              delivery tracking in one place.
            </p>
            <p className="mt-5 font-kannada text-sm text-brand-cream/55">
              ನಿಮ್ಮ ಸ್ಥಳೀಯ ವ್ಯಾಪಾರ ಮಿತ್ರ
            </p>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-cream/55">
              Platform
            </h3>
            <FooterLinkList links={platformLinks} />
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-cream/55">
              Customers
            </h3>
            <FooterLinkList links={customerLinks} />
            <h3 className="mt-7 text-xs font-semibold uppercase tracking-[0.18em] text-brand-cream/55">
              Vendors
            </h3>
            <FooterLinkList links={vendorLinks} />
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-cream/55">
              Contact
            </h3>
            {hasContactEmail ? (
              <a
                href={`mailto:${siteConfig.contactEmail}`}
                className="mt-4 flex items-start gap-2 break-all text-sm text-brand-cream/70 transition-colors hover:text-brand-yellow"
              >
                <Mail className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
                {siteConfig.contactEmail}
              </a>
            ) : (
              <p className="mt-4 text-sm leading-relaxed text-brand-cream/60">
                Public contact details will be published here soon. For now,
                use the contact section for enquiries.
              </p>
            )}
            <a
              href="#contact"
              className="mt-5 flex w-fit items-center gap-1.5 text-sm font-medium text-brand-yellow transition-colors hover:text-brand-cream"
            >
              Contact the team
              <ArrowUpRight className="h-4 w-4" aria-hidden="true" />
            </a>
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-4 border-t border-brand-cream/10 pt-6 text-xs text-brand-cream/60 sm:flex-row sm:items-center sm:justify-between">
          <p>
            © {year} VendorSaathi. All rights reserved.
          </p>
          <div className="flex flex-wrap items-center gap-x-5 gap-y-2">
            <button
              type="button"
              onClick={() => onOpenLegal('privacy')}
              className="transition-colors hover:text-brand-cream"
            >
              Privacy Policy
            </button>
            <button
              type="button"
              onClick={() => onOpenLegal('terms')}
              className="transition-colors hover:text-brand-cream"
            >
              Terms &amp; Conditions
            </button>
            <span className="hidden text-brand-cream/25 sm:inline" aria-hidden="true">
              |
            </span>
            <span>Local commerce, made neighbourly.</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
