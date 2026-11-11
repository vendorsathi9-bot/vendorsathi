import { ArrowUpRight, LockKeyhole, ShieldCheck } from 'lucide-react';

import { ADMIN_CAPABILITIES } from '../../content/site';
import {
  hasAdminDashboard,
  siteConfig,
} from '../../lib/config';
import { Button } from '../ui/Button';
import { Reveal } from '../ui/Reveal';

export function AdminDashboard() {
  return (
    <section
      id="admin"
      aria-labelledby="admin-title"
      className="pb-20 sm:pb-24 lg:pb-28"
    >
      <div className="container-page">
        <Reveal>
          <div className="relative overflow-hidden rounded-4xl border border-brand-border bg-brand-white p-6 shadow-card sm:p-8 lg:p-10">
            <div
              aria-hidden="true"
              className="pointer-events-none absolute -right-20 -top-20 h-56 w-56 rounded-full bg-brand-soft/70"
            />

            <div className="relative grid items-center gap-8 lg:grid-cols-[1.2fr_0.8fr] lg:gap-12">
              <div>
                <div className="flex items-center gap-3">
                  <span className="grid h-10 w-10 place-items-center rounded-xl bg-brand-soft text-brand-espresso">
                    <ShieldCheck className="h-5 w-5" aria-hidden="true" />
                  </span>
                  <p className="text-[0.66rem] font-semibold uppercase tracking-[0.18em] text-brand-brown">
                    For administrators
                  </p>
                </div>

                <h2
                  id="admin-title"
                  className="mt-4 text-2xl font-semibold tracking-[-0.03em] text-brand-espresso sm:text-[1.85rem]"
                >
                  One clear view of platform operations.
                </h2>
                <p className="mt-2.5 max-w-xl text-sm leading-relaxed text-brand-brown">
                  The VendorSaathi admin dashboard gives authorised operators a
                  single place to oversee the marketplace and keep it running
                  smoothly.
                </p>

                <div className="mt-6 grid gap-x-6 gap-y-3 sm:grid-cols-2">
                  {ADMIN_CAPABILITIES.map((capability) => {
                    const Icon = capability.icon;
                    return (
                      <div key={capability.title} className="flex gap-2.5">
                        <Icon
                          className="mt-0.5 h-4 w-4 shrink-0 text-brand-espresso"
                          aria-hidden="true"
                        />
                        <div>
                          <p className="text-[0.82rem] font-semibold text-brand-espresso">
                            {capability.title}
                          </p>
                          <p className="text-[0.78rem] text-brand-brown">
                            {capability.description}
                          </p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              <div className="rounded-3xl border border-brand-border bg-brand-cream p-5 sm:p-6">
                <span className="grid h-10 w-10 place-items-center rounded-xl bg-brand-white text-brand-espresso shadow-sm">
                  <LockKeyhole className="h-5 w-5" aria-hidden="true" />
                </span>
                <h3 className="mt-4 text-base font-semibold text-brand-espresso">
                  Secure operator access
                </h3>
                <p className="mt-1.5 text-[0.82rem] leading-relaxed text-brand-brown">
                  Admin tools are restricted to authorised platform operators.
                </p>

                {hasAdminDashboard ? (
                  <Button
                    href={siteConfig.adminDashboardUrl}
                    external
                    variant="dark"
                    size="sm"
                    className="mt-4 w-full"
                    iconRight={<ArrowUpRight className="h-4 w-4" />}
                  >
                    Open Admin Dashboard
                  </Button>
                ) : (
                  <p className="mt-4 rounded-xl border border-brand-border bg-brand-white px-3.5 py-3 text-[0.74rem] leading-relaxed text-brand-brown">
                    Dashboard access is shared directly with operators. A public
                    sign-in link will be published here when available.
                  </p>
                )}
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
