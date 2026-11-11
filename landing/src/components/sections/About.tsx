import { VALUES } from '../../content/site';
import { Reveal } from '../ui/Reveal';
import { SectionHeading } from '../ui/SectionHeading';

export function About() {
  return (
    <section
      id="about"
      aria-labelledby="about-title"
      className="relative border-y border-brand-border/70 bg-brand-soft/40 py-20 sm:py-24 lg:py-28"
    >
      <div className="container-page">
        <Reveal>
          <SectionHeading
            eyebrow="About VendorSaathi"
            title="Built for the neighbourhood."
            titleId="about-title"
            align="center"
            className="mx-auto"
          />
        </Reveal>

        <Reveal delay={0.06}>
          <div className="mx-auto mt-6 max-w-3xl space-y-4 text-center text-[0.98rem] leading-relaxed text-brand-brown sm:text-[1.05rem]">
            <p>
              VendorSaathi aims to connect local vendors and customers through a
              convenient digital commerce experience — helping local businesses
              establish an online presence while making nearby shopping easier.
            </p>
            <p>
              Instead of pulling attention away from the high street, the
              platform puts the shops already around you on the map: their
              products, their prices, their availability, and their relationship
              with the people they serve.
            </p>
          </div>
        </Reveal>

        <Reveal delay={0.1}>
          <p className="mt-7 text-center font-kannada text-base text-brand-espresso/75 sm:text-lg">
            ನಿಮ್ಮ ಸ್ಥಳೀಯ ವ್ಯಾಪಾರ ಮಿತ್ರ
          </p>
          <p className="mt-1 text-center text-[0.7rem] font-medium uppercase tracking-[0.16em] text-brand-brown/80">
            Your local commerce companion
          </p>
        </Reveal>

        <div className="mx-auto mt-12 grid max-w-4xl gap-4 sm:grid-cols-3">
          {VALUES.map((value, index) => {
            const Icon = value.icon;
            return (
              <Reveal key={value.title} delay={index * 0.06}>
                <div className="h-full rounded-3xl border border-brand-border bg-brand-white p-6 shadow-card">
                  <span className="grid h-10 w-10 place-items-center rounded-xl bg-brand-soft text-brand-espresso">
                    <Icon className="h-[18px] w-[18px]" aria-hidden="true" />
                  </span>
                  <h3 className="mt-4 text-sm font-semibold text-brand-espresso">
                    {value.title}
                  </h3>
                  <p className="mt-1.5 text-[0.82rem] leading-relaxed text-brand-brown">
                    {value.description}
                  </p>
                </div>
              </Reveal>
            );
          })}
        </div>
      </div>
    </section>
  );
}
