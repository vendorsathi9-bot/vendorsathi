import { QRCodeSVG } from 'qrcode.react';
import {
  ArrowRight,
  Clock,
  Download,
  Smartphone,
  Star,
} from 'lucide-react';

import cartScreen from '../../assets/screenshots/cart.webp';
import {
  hasAppDownload,
  siteConfig,
} from '../../lib/config';
import { Button } from '../ui/Button';
import { PhoneFrame } from '../ui/PhoneFrame';
import { Reveal } from '../ui/Reveal';

export function DownloadApp() {
  return (
    <section
      id="download"
      aria-labelledby="download-title"
      className="pb-20 sm:pb-24 lg:pb-28"
    >
      <div className="container-page">
        <Reveal>
          <div className="relative overflow-hidden rounded-[2.5rem] bg-brand-yellow px-6 py-12 shadow-lift sm:px-10 sm:py-14 lg:px-16 lg:py-16">
            <div
              aria-hidden="true"
              className="pointer-events-none absolute inset-0"
            >
              <div className="absolute -right-20 -top-24 h-72 w-72 rounded-full border-[28px] border-brand-espresso/5" />
              <div className="absolute -bottom-24 left-1/3 h-64 w-64 rounded-full bg-brand-cream/20 blur-2xl" />
              <div className="dot-grid absolute bottom-8 left-6 h-24 w-32 opacity-25" />
            </div>

            <div className="relative grid items-center gap-12 lg:grid-cols-[1.08fr_0.92fr] lg:gap-8">
              <div>
                <span className="inline-flex items-center gap-2 rounded-full border border-brand-espresso/15 bg-brand-cream/40 px-3.5 py-1.5 text-[0.66rem] font-semibold uppercase tracking-[0.16em] text-brand-espresso/75">
                  <Smartphone className="h-3.5 w-3.5" aria-hidden="true" />
                  Get the app
                </span>

                <h2
                  id="download-title"
                  className="mt-4 max-w-xl text-display-sm font-semibold text-brand-espresso"
                >
                  Take Your Local Market Everywhere.
                </h2>

                <p className="mt-4 max-w-xl text-pretty text-[0.98rem] leading-relaxed text-brand-espresso/75 sm:text-[1.05rem]">
                  Nearby stores, your cart, loyalty rewards, and live delivery
                  tracking — all in one Android app built for everyday local
                  shopping.
                </p>

                {hasAppDownload ? (
                  <div className="mt-7 flex flex-col gap-5 sm:flex-row sm:items-center">
                    <Button
                      href={siteConfig.appDownloadUrl}
                      external
                      variant="dark"
                      size="lg"
                      icon={<Download className="h-[18px] w-[18px]" />}
                      iconRight={<ArrowRight className="h-[18px] w-[18px]" />}
                    >
                      Download for Android
                    </Button>

                    <div className="flex items-center gap-3 rounded-2xl border border-brand-espresso/10 bg-brand-cream p-3">
                      <div className="rounded-xl bg-brand-white p-1.5">
                        <QRCodeSVG
                          value={siteConfig.appDownloadUrl}
                          size={86}
                          bgColor="#FFFFFF"
                          fgColor="#38200F"
                          level="M"
                          role="img"
                          aria-label="QR code for the VendorSaathi Android app download"
                        />
                      </div>
                      <p className="max-w-[120px] text-[0.72rem] font-medium leading-snug text-brand-espresso/70">
                        Scan to download on your Android phone.
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="mt-7">
                    <span className="inline-flex items-center gap-2 rounded-full bg-brand-espresso px-4 py-2 text-xs font-semibold text-brand-cream">
                      <Clock className="h-3.5 w-3.5" aria-hidden="true" />
                      Coming soon
                    </span>
                    <p className="mt-3 max-w-md text-sm leading-relaxed text-brand-espresso/70">
                      The public Android release is being prepared. Download
                      details will appear here as soon as the app is available.
                    </p>
                    <Button
                      href="#customers"
                      variant="dark"
                      className="mt-5"
                      iconRight={<ArrowRight className="h-[17px] w-[17px]" />}
                    >
                      Explore the experience
                    </Button>
                  </div>
                )}
              </div>

              <div className="relative mx-auto w-full max-w-[340px]">
                <div
                  aria-hidden="true"
                  className="absolute inset-x-8 bottom-4 top-8 rotate-[-3deg] rounded-[3rem] bg-brand-cream/30"
                />
                <div className="relative mx-auto w-[72%] max-w-[250px]">
                  <PhoneFrame
                    src={cartScreen}
                    alt="VendorSaathi cart screen showing items, totals, and checkout"
                    className="ring-brand-espresso/10"
                  />
                </div>

                <div className="absolute bottom-1 -right-1 z-10 flex max-w-[206px] items-center gap-3 rounded-2xl border border-brand-espresso/10 bg-brand-cream p-3 shadow-float motion-safe:animate-float sm:right-2 lg:-right-3">
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-white">
                    <Star
                      className="h-[18px] w-[18px] text-brand-espresso"
                      aria-hidden="true"
                    />
                  </span>
                  <span>
                    <span className="block text-[0.66rem] font-medium uppercase tracking-[0.08em] text-brand-brown">
                      Loyalty rewards
                    </span>
                    <span className="mt-0.5 block text-[0.78rem] font-semibold leading-tight text-brand-espresso">
                      Points on completed orders
                    </span>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
