import { cn } from '../../lib/cn';

export function StorefrontMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      className={className}
    >
      <path d="M21.9 8.89l-1.05-4.37c-.22-.9-1-1.52-1.91-1.52H5.05c-.9 0-1.69.62-1.91 1.52L2.1 8.89c-.24 1.02-.02 2.06.62 2.88.08.11.19.19.28.29V19c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2v-6.94c.09-.09.2-.18.28-.28.64-.82.87-1.87.62-2.89zm-2.99-5.26l1.05 4.37c.1.42.01.84-.25 1.17-.14.18-.44.47-.94.47-.61 0-1.14-.49-1.21-1.14L16.98 5l1.93-1.37zM13 5h1.96l.54 4.52c.05.39-.07.78-.33 1.07-.22.26-.54.41-.95.41-.67 0-1.22-.59-1.22-1.31V5zM8.15 5h1.96v4.69c0 .72-.55 1.31-1.29 1.31-.34 0-.65-.15-.89-.41a1.42 1.42 0 0 1-.33-1.07L8.15 5zm-4.45.26L4.75 9.2c-.08.69-.59 1.14-1.21 1.14-.5 0-.8-.29-.94-.47-.25-.33-.35-.75-.25-1.17l1.05-4.37L5.34 5.26zM4 19v-6.03c.3.05.62.08.95.08.97 0 1.82-.35 2.49-.95.74.6 1.64.95 2.71.95.97 0 1.85-.35 2.54-.95.71.6 1.6.95 2.6.95.97 0 1.86-.35 2.55-.95.66.6 1.51.95 2.49.95.32 0 .64-.03.94-.08V19H4z" />
    </svg>
  );
}

interface LogoProps {
  tone?: 'dark' | 'light';
  className?: string;
  markClassName?: string;
  wordmarkClassName?: string;
}

export function Logo({
  tone = 'dark',
  className,
  markClassName,
  wordmarkClassName,
}: LogoProps) {
  return (
    <span className={cn('inline-flex items-center gap-2.5', className)}>
      <span
        className={cn(
          'grid h-10 w-10 shrink-0 place-items-center rounded-[13px] bg-brand-yellow shadow-[inset_0_-2px_0_rgba(56,32,15,0.16)]',
          markClassName,
        )}
      >
        <StorefrontMark className="h-[22px] w-[22px] text-brand-espresso" />
      </span>
      <span
        className={cn(
          'text-[1.16rem] font-semibold leading-none tracking-[-0.035em]',
          tone === 'light' ? 'text-brand-cream' : 'text-brand-espresso',
          wordmarkClassName,
        )}
      >
        VendorSaathi
      </span>
    </span>
  );
}
