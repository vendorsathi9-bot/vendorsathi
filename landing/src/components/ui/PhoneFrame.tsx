import { cn } from '../../lib/cn';

interface PhoneFrameProps {
  src: string;
  alt: string;
  className?: string;
  imageClassName?: string;
  priority?: boolean;
}

export function PhoneFrame({
  src,
  alt,
  className,
  imageClassName,
  priority = false,
}: PhoneFrameProps) {
  return (
    <div
      className={cn(
        'relative rounded-[2.65rem] bg-brand-espresso p-[9px] shadow-phone ring-1 ring-brand-espresso/10',
        className,
      )}
    >
      <div className="relative overflow-hidden rounded-[2.08rem] bg-brand-cream">
        <img
          src={src}
          alt={alt}
          width={720}
          height={1616}
          loading={priority ? 'eager' : 'lazy'}
          decoding="async"
          className={cn('block h-auto w-full', imageClassName)}
        />
        <span
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 rounded-[2.08rem] ring-1 ring-inset ring-brand-espresso/5"
        />
      </div>
    </div>
  );
}
