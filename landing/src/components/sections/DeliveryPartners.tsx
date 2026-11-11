import {
  ArrowRight,
  Bike,
  CircleCheck,
  ClipboardList,
  Navigation,
  Store,
} from 'lucide-react';

import { DELIVERY_POINTS } from '../../content/site';
import { Button } from '../ui/Button';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

const workflow = [
  {
    title: 'Assigned',
    description: 'A delivery request appears with pickup and drop-off details.',
    icon: ClipboardList,
  },
  {
    title: 'Pickup',
    description: 'Collect the order from the vendor.',
    icon: Store,
  },
  {
    title: 'On the way',
    description: 'Follow the route while GPS sharing keeps the customer updated.',
    icon: Navigation,
  },
  {
    title: 'Delivered',
    description: 'Mark the delivery complete and close the assignment.',
    icon: CircleCheck,
  },
];

export function DeliveryPartners() {
  return (
    <section
      id="delivery"
      aria-labelledby="delivery-title"
      className="relative py-20 sm:py-24 lg:py-28"
    >
      <div className="container-page">
        <div className="grid items-center gap-14 lg:grid-cols-[1.02fr_0.98fr] lg:gap-16">
          <div>
            <Reveal>
              <SectionHeading
                eyebrow="Delivery partners"
                title="Every delivery, clearly in view."
                titleId="delivery-title"
                description="Delivery partners get the details they need, a simple pickup-to-drop-off flow, and a direct way to keep customers informed."
              />
            </Reveal>

            <ul className="mt-8 grid gap-5 sm:grid-cols-2">
              {DELIVERY_POINTS.map((point, index) => {
                const Icon = point.icon;
                return (
                  <Reveal
                    key={point.title}
                    as="li"
                    delay={index * 0.05}
                    y={14}
                    className="rounded-2xl border border-brand-border bg-brand-white p-5 shadow-card"
                  >
                    <span className="grid h-9 w-9 place-items-center rounded-xl bg-brand-soft text-brand-espresso">
                      <Icon className="h-4 w-4" aria-hidden="true" />
                    </span>
                    <h3 className="mt-3 text-sm font-semibold text-brand-espresso">
                      {point.title}
                    </h3>
                    <p className="mt-1 text-[0.82rem] leading-relaxed text-brand-brown">
                      {point.description}
                    </p>
                  </Reveal>
                );
              })}
            </ul>

            <Reveal delay={0.12}>
              <Button
                href="#contact"
                variant="dark"
                className="mt-8"
                icon={<Bike className="h-[17px] w-[17px]" />}
                iconRight={<ArrowRight className="h-[17px] w-[17px]" />}
              >
                Deliver with VendorSaathi
              </Button>
            </Reveal>
          </div>

          <Reveal delay={0.1}>
            <div className="relative overflow-hidden rounded-4xl border border-brand-border bg-brand-white p-6 shadow-card sm:p-8">
              <div
                aria-hidden="true"
                className="pointer-events-none absolute -right-16 -top-16 h-48 w-48 rounded-full bg-brand-soft/70"
              />

              <div className="relative flex items-start justify-between gap-4">
                <div>
                  <p className="text-[0.64rem] font-semibold uppercase tracking-[0.18em] text-brand-brown">
                    Delivery workflow
                  </p>
                  <h3 className="mt-1.5 text-lg font-semibold tracking-[-0.02em] text-brand-espresso">
                    Assigned to delivered
                  </h3>
                </div>
                <span className="rounded-full border border-brand-border bg-brand-cream px-3 py-1 text-[0.6rem] font-semibold uppercase tracking-[0.12em] text-brand-brown">
                  Illustrative
                </span>
              </div>

              <div className="relative mt-7">
                <span
                  aria-hidden="true"
                  className="absolute bottom-3 left-[19px] top-3 border-l border-dashed border-brand-border"
                />
                <ol className="relative space-y-6">
                  {workflow.map((step, index) => {
                    const Icon = step.icon;
                    const isLast = index === workflow.length - 1;
                    return (
                      <li key={step.title} className="relative flex gap-4">
                        <span
                          className={`z-10 grid h-10 w-10 shrink-0 place-items-center rounded-full border ${
                            isLast
                              ? 'border-success/25 bg-success-soft text-success'
                              : 'border-brand-border bg-brand-cream text-brand-espresso'
                          }`}
                        >
                          <Icon
                            className="h-[18px] w-[18px]"
                            aria-hidden="true"
                          />
                        </span>
                        <div className="pt-1.5">
                          <h4 className="text-sm font-semibold text-brand-espresso">
                            {step.title}
                          </h4>
                          <p className="mt-0.5 text-[0.82rem] leading-relaxed text-brand-brown">
                            {step.description}
                          </p>
                        </div>
                      </li>
                    );
                  })}
                </ol>
              </div>

              <div className="mt-7 flex items-center justify-between gap-3 rounded-2xl bg-brand-soft px-4 py-3">
                <span className="flex items-center gap-2 text-sm font-medium text-brand-espresso">
                  <Navigation className="h-4 w-4" aria-hidden="true" />
                  Live GPS sharing
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-white px-2.5 py-1 text-[0.62rem] font-semibold uppercase tracking-[0.1em] text-success">
                  <span
                    className="h-1.5 w-1.5 rounded-full bg-success"
                    aria-hidden="true"
                  />
                  On
                </span>
              </div>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
