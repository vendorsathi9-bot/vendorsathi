import {
  ArrowRight,
  Bike,
  Mail,
  MessageSquare,
  ShoppingBag,
  Store,
} from 'lucide-react';

import {
  contactHref,
  hasContactEmail,
  siteConfig,
} from '../../lib/config';
import { Button } from '../ui/Button';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

const audiences = [
  {
    title: 'Local vendors',
    description:
      'Bring your storefront online, manage your catalogue, and handle orders digitally.',
    icon: Store,
  },
  {
    title: 'Delivery partners',
    description:
      'Take on delivery assignments with clear pickup, drop-off, and status flows.',
    icon: Bike,
  },
  {
    title: 'Customers',
    description:
      'Questions about a nearby store, an order, or the customer app? Get in touch.',
    icon: ShoppingBag,
  },
];

export function Contact() {
  return (
    <section
      id="contact"
      aria-labelledby="contact-title"
      className="relative py-20 sm:py-24 lg:py-28"
    >
      <div className="container-page">
        <div className="grid gap-12 lg:grid-cols-[1fr_0.92fr] lg:gap-16">
          <div>
            <Reveal>
              <SectionHeading
                eyebrow="Contact"
                title="Let’s make local commerce easier."
                titleId="contact-title"
                description="Whether you run a shop, deliver in your neighbourhood, or shop locally, we would like to hear from you."
              />
            </Reveal>

            <ul className="mt-8 divide-y divide-brand-border border-y border-brand-border">
              {audiences.map((audience, index) => {
                const Icon = audience.icon;
                return (
                  <Reveal
                    key={audience.title}
                    as="li"
                    delay={index * 0.05}
                    y={12}
                    className="flex gap-4 py-5"
                  >
                    <span className="mt-0.5 grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-brand-soft text-brand-espresso">
                      <Icon className="h-[18px] w-[18px]" aria-hidden="true" />
                    </span>
                    <div>
                      <h3 className="text-sm font-semibold text-brand-espresso">
                        {audience.title}
                      </h3>
                      <p className="mt-1 text-sm leading-relaxed text-brand-brown">
                        {audience.description}
                      </p>
                    </div>
                  </Reveal>
                );
              })}
            </ul>
          </div>

          <Reveal delay={0.1}>
            <div className="relative overflow-hidden rounded-4xl border border-brand-border bg-brand-white p-6 shadow-card sm:p-8">
              <div
                aria-hidden="true"
                className="pointer-events-none absolute -right-16 -top-16 h-48 w-48 rounded-full bg-brand-soft/70"
              />

              <div className="relative">
                <span className="grid h-11 w-11 place-items-center rounded-2xl bg-brand-yellow text-brand-espresso">
                  <MessageSquare className="h-5 w-5" aria-hidden="true" />
                </span>
                <h3 className="mt-5 text-xl font-semibold tracking-[-0.025em] text-brand-espresso">
                  Start a conversation
                </h3>

                {hasContactEmail ? (
                  <>
                    <p className="mt-2 text-sm leading-relaxed text-brand-brown">
                      Write to the VendorSaathi team and we will route your
                      enquiry to the right person.
                    </p>

                    <a
                      href={`mailto:${siteConfig.contactEmail}`}
                      className="mt-5 inline-flex max-w-full items-center gap-2 break-all text-sm font-semibold text-brand-espresso underline decoration-brand-yellow decoration-2 underline-offset-4 hover:text-brand-brown"
                    >
                      <Mail className="h-4 w-4 shrink-0" aria-hidden="true" />
                      {siteConfig.contactEmail}
                    </a>

                    <div className="mt-6 space-y-2">
                      {[
                        { label: 'Vendor onboarding', subject: 'Vendor onboarding' },
                        {
                          label: 'Delivery partner onboarding',
                          subject: 'Delivery partner onboarding',
                        },
                        { label: 'Customer support', subject: 'Customer support' },
                      ].map((item) => (
                        <a
                          key={item.label}
                          href={contactHref(item.subject)}
                          className="group flex items-center justify-between rounded-xl border border-brand-border bg-brand-cream px-4 py-3 text-sm font-medium text-brand-espresso transition hover:border-brand-yellow/70 hover:bg-brand-soft/60"
                        >
                          {item.label}
                          <ArrowRight
                            className="h-4 w-4 text-brand-brown transition-transform group-hover:translate-x-0.5"
                            aria-hidden="true"
                          />
                        </a>
                      ))}
                    </div>
                  </>
                ) : (
                  <>
                    <p className="mt-2 text-sm leading-relaxed text-brand-brown">
                      Public contact details will be published here soon. Vendor
                      onboarding, delivery partnership, and customer enquiries
                      are being routed through the VendorSaathi team in the
                      meantime.
                    </p>

                    <div className="mt-5 rounded-2xl border border-brand-border bg-brand-cream p-4">
                      <p className="flex items-start gap-2 text-xs leading-relaxed text-brand-brown">
                        <Mail
                          className="mt-0.5 h-4 w-4 shrink-0 text-brand-espresso"
                          aria-hidden="true"
                        />
                        A dedicated contact address and phone line are coming
                        with the public launch.
                      </p>
                    </div>
                  </>
                )}

                <div className="mt-6 border-t border-brand-border pt-5">
                  <p className="text-[0.66rem] font-semibold uppercase tracking-[0.16em] text-brand-brown">
                    While you’re here
                  </p>
                  <div className="mt-3 flex flex-wrap gap-2.5">
                    <Button href="#download" size="sm" variant="secondary">
                      Download the app
                    </Button>
                    <Button href="#admin" size="sm" variant="ghost">
                      Admin overview
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
