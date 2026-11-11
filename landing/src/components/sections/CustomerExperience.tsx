import { Languages } from 'lucide-react';

import ordersScreen from '../../assets/screenshots/orders.webp';
import storefrontScreen from '../../assets/screenshots/storefront.webp';
import { CUSTOMER_HIGHLIGHTS } from '../../content/site';
import { PhoneFrame } from '../ui/PhoneFrame';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

export function CustomerExperience() {
  return (
    <section
      id="customers"
      aria-labelledby="customers-title"
      className="relative overflow-hidden border-y border-brand-border/70 bg-brand-soft/40 py-20 sm:py-24 lg:py-28"
    >
      <div className="container-page">
        <div className="grid items-center gap-14 lg:grid-cols-[0.95fr_1.05fr] lg:gap-16">
          <div>
            <Reveal>
              <SectionHeading
                eyebrow="Customer experience"
                title="A calmer way to shop nearby."
                titleId="customers-title"
                description="From the first search to the final delivery, every step is designed to feel clear, local, and familiar."
              />
            </Reveal>

            <ul className="mt-8 space-y-5">
              {CUSTOMER_HIGHLIGHTS.map((item, index) => {
                const Icon = item.icon;
                return (
                  <Reveal
                    key={item.title}
                    as="li"
                    delay={index * 0.04}
                    y={14}
                    className="flex gap-3.5"
                  >
                    <span className="mt-0.5 grid h-9 w-9 shrink-0 place-items-center rounded-xl border border-brand-border bg-brand-white text-brand-espresso">
                      <Icon className="h-4 w-4" aria-hidden="true" />
                    </span>
                    <div>
                      <h3 className="text-sm font-semibold text-brand-espresso">
                        {item.title}
                      </h3>
                      <p className="mt-0.5 text-sm leading-relaxed text-brand-brown">
                        {item.description}
                      </p>
                    </div>
                  </Reveal>
                );
              })}
            </ul>
          </div>

          <Reveal delay={0.12} className="relative">
            <div
              aria-hidden="true"
              className="absolute inset-x-8 bottom-10 top-12 rounded-[3rem] bg-brand-yellow/15"
            />
            <div
              aria-hidden="true"
              className="dot-grid absolute -right-2 top-0 h-24 w-24 opacity-40"
            />

            <div className="relative mx-auto flex max-w-[520px] items-start gap-4 sm:gap-5">
              <figure className="w-[53%] shrink-0">
                <PhoneFrame
                  src={storefrontScreen}
                  alt="VendorSaathi storefront screen showing local products, prices, and cart controls"
                  className="transition-transform duration-300 hover:-translate-y-1"
                />
                <figcaption className="mt-3 text-center text-[0.7rem] font-medium text-brand-brown">
                  Browse a local storefront
                </figcaption>
              </figure>

              <figure className="mt-10 w-[44%] shrink-0 sm:mt-12">
                <PhoneFrame
                  src={ordersScreen}
                  alt="VendorSaathi orders screen showing order status and tracking"
                  className="transition-transform duration-300 hover:-translate-y-1"
                />
                <figcaption className="mt-3 text-center text-[0.7rem] font-medium text-brand-brown">
                  Follow every order
                </figcaption>
              </figure>
            </div>

            <div className="relative mt-7 flex justify-center">
              <span className="inline-flex items-center gap-2 rounded-full border border-brand-border bg-brand-white px-4 py-2 text-xs font-medium text-brand-brown shadow-sm">
                <Languages
                  className="h-4 w-4 text-brand-espresso"
                  aria-hidden="true"
                />
                Available in English and{' '}
                <span className="font-kannada">ಕನ್ನಡ</span>
              </span>
            </div>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
