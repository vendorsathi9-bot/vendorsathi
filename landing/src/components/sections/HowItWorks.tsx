import { STEPS } from '../../content/site';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

export function HowItWorks() {
  return (
    <section
      id="how-it-works"
      aria-labelledby="how-it-works-title"
      className="relative border-y border-brand-border/70 bg-brand-soft/45 py-20 sm:py-24 lg:py-28"
    >
      <div className="container-page">
        <Reveal>
          <SectionHeading
            eyebrow="How it works"
            title="From your street to your doorstep."
            titleId="how-it-works-title"
            align="center"
            className="mx-auto"
            description="One simple path from finding a neighbourhood store to receiving your order — with clear status at every step."
          />
        </Reveal>

        <div className="relative mt-14 grid gap-5 md:grid-cols-3">
          <div
            aria-hidden="true"
            className="absolute left-[17%] right-[17%] top-[54px] hidden border-t border-dashed border-brand-espresso/15 md:block"
          />

          {STEPS.map((step, index) => {
            const Icon = step.icon;
            return (
              <Reveal key={step.title} delay={index * 0.08}>
                <article className="group relative flex h-full flex-col rounded-3xl border border-brand-border bg-brand-white p-6 shadow-card transition-all duration-300 hover:-translate-y-1 hover:border-brand-yellow/70 hover:shadow-lift sm:p-7">
                  <div className="flex items-start justify-between">
                    <span className="grid h-12 w-12 place-items-center rounded-2xl bg-brand-soft text-brand-espresso transition-colors duration-300 group-hover:bg-brand-yellow">
                      <Icon className="h-[22px] w-[22px]" aria-hidden="true" />
                    </span>
                    <span
                      aria-hidden="true"
                      className="text-4xl font-semibold tracking-[-0.05em] text-brand-espresso/10"
                    >
                      0{index + 1}
                    </span>
                  </div>

                  <h3 className="mt-6 text-lg font-semibold tracking-[-0.02em] text-brand-espresso">
                    {step.title}
                  </h3>
                  <p className="mt-2.5 text-sm leading-relaxed text-brand-brown">
                    {step.description}
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
