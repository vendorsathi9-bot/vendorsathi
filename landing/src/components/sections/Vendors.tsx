import { ArrowRight, BookOpen } from 'lucide-react';

import vendorProductsScreen from '../../assets/screenshots/vendor-products.webp';
import { VENDOR_BENEFITS } from '../../content/site';
import { Button } from '../ui/Button';
import { PhoneFrame } from '../ui/PhoneFrame';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

export function Vendors() {
  return (
    <section
      id="vendors"
      aria-labelledby="vendors-title"
      className="relative overflow-hidden bg-brand-espresso py-20 text-brand-cream sm:py-24 lg:py-28"
    >
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <div className="absolute -right-24 -top-24 h-80 w-80 rounded-full bg-brand-yellow/10 blur-3xl" />
        <div className="absolute -bottom-24 left-1/4 h-72 w-72 rounded-full bg-brand-yellow/5 blur-3xl" />
        <div className="dot-grid-cream absolute right-10 top-16 hidden h-32 w-48 opacity-60 lg:block" />
      </div>

      <div className="container-page relative">
        <div className="grid items-center gap-14 lg:grid-cols-[1.06fr_0.94fr] lg:gap-16">
          <div>
            <Reveal>
              <SectionHeading
                eyebrow="For local vendors"
                title="Your shop, online — without the chaos."
                titleId="vendors-title"
                tone="light"
                description="VendorSaathi gives neighbourhood businesses a simple digital counter: a storefront, a catalogue, an order queue, and the numbers that matter."
              />
            </Reveal>

            <div className="mt-9 grid gap-x-6 gap-y-6 sm:grid-cols-2">
              {VENDOR_BENEFITS.map((benefit, index) => {
                const Icon = benefit.icon;
                return (
                  <Reveal key={benefit.title} delay={index * 0.04} y={14}>
                    <div className="flex gap-3">
                      <span className="mt-0.5 grid h-9 w-9 shrink-0 place-items-center rounded-xl border border-brand-cream/10 bg-brand-cream/5 text-brand-yellow">
                        <Icon className="h-4 w-4" aria-hidden="true" />
                      </span>
                      <div>
                        <h3 className="text-sm font-semibold text-brand-cream">
                          {benefit.title}
                        </h3>
                        <p className="mt-1 text-[0.82rem] leading-relaxed text-brand-cream/60">
                          {benefit.description}
                        </p>
                      </div>
                    </div>
                  </Reveal>
                );
              })}
            </div>

            <Reveal delay={0.12}>
              <div className="mt-10 flex flex-wrap items-center gap-4">
                <Button
                  href="#contact"
                  iconRight={<ArrowRight className="h-[17px] w-[17px]" />}
                >
                  Join VendorSaathi
                </Button>
                <p className="text-xs text-brand-cream/50">
                  Vendor onboarding is coordinated with the VendorSaathi team.
                </p>
              </div>
            </Reveal>
          </div>

          <Reveal delay={0.1} className="relative mx-auto w-full max-w-[390px]">
            <div
              aria-hidden="true"
              className="absolute inset-x-10 bottom-6 top-8 rotate-[-3deg] rounded-[3rem] border border-brand-cream/10 bg-brand-cream/5"
            />
            <div className="relative mx-auto w-[76%] max-w-[280px]">
              <PhoneFrame
                src={vendorProductsScreen}
                alt="VendorSaathi vendor portal showing the product catalogue with availability toggles"
                className="ring-brand-cream/10"
              />
            </div>

            <div className="absolute -left-1 bottom-0 z-10 flex max-w-[210px] items-center gap-3 rounded-2xl border border-brand-border bg-brand-white p-3 shadow-float motion-safe:animate-float sm:left-2 lg:-left-12">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-soft">
                <BookOpen
                  className="h-[18px] w-[18px] text-brand-espresso"
                  aria-hidden="true"
                />
              </span>
              <span>
                <span className="block text-[0.66rem] font-medium uppercase tracking-[0.08em] text-brand-brown">
                  Udhaar ledger
                </span>
                <span className="mt-0.5 block text-[0.78rem] font-semibold leading-tight text-brand-espresso">
                  Credit, recorded clearly
                </span>
              </span>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
