import type { ReactNode } from 'react';
import { cn } from '../../lib/cn';

interface SectionHeadingProps {
  eyebrow: string;
  title: ReactNode;
  description?: ReactNode;
  align?: 'left' | 'center';
  tone?: 'dark' | 'light';
  className?: string;
  titleId?: string;
}

export function SectionHeading({
  eyebrow,
  title,
  description,
  align = 'left',
  tone = 'dark',
  className,
  titleId,
}: SectionHeadingProps) {
  const isLight = tone === 'light';

  return (
    <div
      className={cn(
        'flex flex-col',
        align === 'center' ? 'items-center text-center' : 'items-start',
        className,
      )}
    >
      <span
        className={cn(
          'inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-[0.66rem] font-semibold uppercase tracking-[0.18em]',
          isLight
            ? 'border-brand-cream/15 bg-brand-cream/5 text-brand-cream/70'
            : 'border-brand-border bg-brand-white/70 text-brand-brown',
        )}
      >
        <span className="h-1.5 w-1.5 rounded-full bg-brand-yellow" aria-hidden="true" />
        {eyebrow}
      </span>

      <h2
        id={titleId}
        className={cn(
          'mt-4 max-w-3xl text-display-sm font-semibold',
          isLight ? 'text-brand-cream' : 'text-brand-espresso',
        )}
      >
        {title}
      </h2>

      {description ? (
        <p
          className={cn(
            'mt-4 max-w-2xl text-pretty text-[0.98rem] leading-relaxed sm:text-[1.05rem]',
            isLight ? 'text-brand-cream/70' : 'text-brand-brown',
          )}
        >
          {description}
        </p>
      ) : null}
    </div>
  );
}
