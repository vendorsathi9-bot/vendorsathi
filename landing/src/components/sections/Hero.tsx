import { motion, useReducedMotion } from 'framer-motion';
import {
  ArrowRight,
  Check,
  MapPinned,
  Navigation,
  Play,
  Smartphone,
} from 'lucide-react';

import nearbyVendors from '../../assets/screenshots/nearby-vendors.webp';
import { Button } from '../ui/Button';
import { PhoneFrame } from '../ui/PhoneFrame';

export function Hero() {
  const reduceMotion = useReducedMotion();

  const entrance = (delay: number) => ({
    initial: reduceMotion ? false : { opacity: 0, y: 26 },
    animate: reduceMotion ? undefined : { opacity: 1, y: 0 },
    transition: {
      duration: 0.7,
      delay,
      ease: [0.22, 1, 0.36, 1] as const,
    },
  });

  return (
    <section
      id="home"
      aria-labelledby="hero-title"
      className="relative overflow-hidden pb-20 pt-10 sm:pb-24 sm:pt-16 lg:pb-28 lg:pt-20"
    >
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <div className="absolute -right-28 -top-32 h-[420px] w-[420px] rounded-full bg-brand-soft/80" />
        <div className="absolute right-[8%] top-16 hidden h-[500px] w-[500px] rounded-full border border-brand-yellow/20 lg:block" />
        <div className="absolute -left-16 top-1/3 h-56 w-56 rounded-full bg-brand-soft/50 blur-3xl" />
        <div className="dot-grid absolute bottom-12 left-[44%] hidden h-28 w-40 opacity-40 lg:block" />
      </div>

      <div className="container-page relative">
        <div className="grid items-center gap-14 lg:grid-cols-[1.02fr_0.98fr] lg:gap-8">
          <div>
            <motion.div {...entrance(0.05)}>
              <span className="inline-flex items-center gap-2 rounded-full border border-brand-border bg-brand-white/80 px-3.5 py-1.5 text-[0.66rem] font-semibold uppercase tracking-[0.16em] text-brand-brown shadow-sm">
                <span
                  className="h-1.5 w-1.5 rounded-full bg-brand-yellow"
                  aria-hidden="true"
                />
                Local commerce, made neighbourly
              </span>
            </motion.div>

            <motion.h1
              id="hero-title"
              {...entrance(0.12)}
              className="mt-5 text-display font-semibold text-brand-espresso"
            >
              Your Local Market,
              <span className="relative mt-1 block w-fit">
                <span
                  aria-hidden="true"
                  className="absolute inset-x-0 bottom-[0.1em] h-[0.26em] -rotate-[0.7deg] rounded-sm bg-brand-yellow/40"
                />
                <span className="relative">Just a Tap Away.</span>
              </span>
            </motion.h1>

            <motion.p
              {...entrance(0.2)}
              className="mt-6 max-w-xl text-pretty text-base leading-relaxed text-brand-brown sm:text-lg"
            >
              Discover nearby stores, explore local products, and get everything
              you need delivered conveniently with VendorSaathi.
            </motion.p>

            <motion.div
              {...entrance(0.28)}
              className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center"
            >
              <Button
                href="#download"
                size="lg"
                icon={<Smartphone className="h-[18px] w-[18px]" />}
                iconRight={<ArrowRight className="h-[18px] w-[18px]" />}
              >
                Get the App
              </Button>
              <Button
                href="#how-it-works"
                variant="secondary"
                size="lg"
                icon={<Play className="h-[17px] w-[17px]" />}
              >
                Explore How It Works
              </Button>
            </motion.div>

            <motion.ul
              {...entrance(0.36)}
              className="mt-9 flex flex-wrap gap-x-6 gap-y-3 text-[0.82rem] font-medium text-brand-brown"
            >
              {[
                'Nearby discovery',
                'Live order tracking',
                'English & Kannada',
              ].map((item) => (
                <li key={item} className="flex items-center gap-1.5">
                  <Check
                    className="h-4 w-4 text-brand-espresso"
                    strokeWidth={2.6}
                    aria-hidden="true"
                  />
                  {item}
                </li>
              ))}
            </motion.ul>
          </div>

          <motion.div
            initial={
              reduceMotion ? false : { opacity: 0, y: 28, scale: 0.97 }
            }
            animate={
              reduceMotion ? undefined : { opacity: 1, y: 0, scale: 1 }
            }
            transition={{
              duration: 0.8,
              delay: 0.18,
              ease: [0.22, 1, 0.36, 1],
            }}
            className="relative mx-auto w-full max-w-[440px] lg:max-w-none"
          >
            <div
              aria-hidden="true"
              className="absolute inset-x-8 bottom-4 top-10 rotate-3 rounded-[3rem] border border-brand-border bg-brand-soft/70"
            />
            <div
              aria-hidden="true"
              className="dot-grid absolute -right-1 top-2 h-24 w-24 opacity-50"
            />
            <div
              aria-hidden="true"
              className="absolute -left-3 bottom-16 h-16 w-16 rounded-full border-[9px] border-brand-yellow/35"
            />

            <div className="relative mx-auto w-[74%] max-w-[286px] sm:w-[68%] lg:w-[298px]">
              <PhoneFrame
                src={nearbyVendors}
                alt="VendorSaathi customer app showing nearby vendors sorted by distance"
                priority
                className="rotate-[1.4deg] transition-transform duration-500 ease-out hover:rotate-0"
              />
            </div>

            <div className="absolute left-0 top-[19%] z-10 flex max-w-[196px] items-center gap-3 rounded-2xl border border-brand-border bg-brand-white/95 p-3 shadow-float motion-safe:animate-float sm:left-2 lg:-left-2">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-soft">
                <MapPinned
                  className="h-[18px] w-[18px] text-brand-espresso"
                  aria-hidden="true"
                />
              </span>
              <span>
                <span className="block text-[0.66rem] font-medium uppercase tracking-[0.08em] text-brand-brown">
                  Nearby vendors
                </span>
                <span className="mt-0.5 block text-[0.78rem] font-semibold leading-tight text-brand-espresso">
                  Shops around you
                </span>
              </span>
            </div>

            <div className="absolute -bottom-1 right-0 z-10 flex max-w-[206px] items-center gap-3 rounded-2xl border border-brand-border bg-brand-white/95 p-3 shadow-float motion-safe:animate-float-slow sm:right-2 lg:-right-3">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-soft">
                <Navigation
                  className="h-[18px] w-[18px] text-brand-espresso"
                  aria-hidden="true"
                />
              </span>
              <span>
                <span className="flex items-center gap-1.5 text-[0.66rem] font-medium uppercase tracking-[0.08em] text-brand-brown">
                  Live tracking
                  <span
                    className="h-1.5 w-1.5 rounded-full bg-success"
                    aria-hidden="true"
                  />
                </span>
                <span className="mt-0.5 block text-[0.78rem] font-semibold leading-tight text-brand-espresso">
                  Follow your delivery
                </span>
              </span>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
