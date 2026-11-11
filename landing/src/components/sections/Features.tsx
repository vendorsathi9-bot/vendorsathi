import { FEATURES } from '../../content/site';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

export function Features() {
  return (
    <section
      id="features"
      aria-labelledby="features-title"
      className="relative py-20 sm:py-24 lg:py-28"
    >
      <div
        aria-hidden="true"
        className="pointer-events-none absolute right-0 top-24 h-64 w-64 rounded-full bg-brand-soft/60 blur-3xl"
      />

      <div className="container-page relative">
        <Reveal>
          <SectionHeading
            eyebrow="Features"
            title="Everything a neighbourhood needs, in one place."
            titleId="features-title"
            className="mx-auto"
            align="center"
            description="Purpose-built tools for customers, vendors, and delivery partners — without unnecessary complexity."
          />
        </Reveal>

        <div className="mt-14 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {FEATURES.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <Reveal key={feature.title} delay={(index % 4) * 0.05}>
                <article className="group h-full rounded-3xl border border-brand-border bg-brand-white p-6 shadow-card transition-all duration-300 hover:-translate-y-1 hover:border-brand-yellow/60 hover:shadow-lift">
                  <span className="grid h-11 w-11 place-items-center rounded-2xl bg-brand-soft text-brand-espresso transition-colors duration-300 group-hover:bg-brand-yellow">
                    <Icon className="h-5 w-5" aria-hidden="true" />
                  </span>
                  <h3 className="mt-5 text-[1.02rem] font-semibold tracking-[-0.015em] text-brand-espresso">
                    {feature.title}
                  </h3>
                  <p className="mt-2 text-sm leading-relaxed text-brand-brown">
                    {feature.description}
                  </p>
                </article>
              </Reveal>
            );
          })}
        </div>
      </div>
    </section>
  );
}
