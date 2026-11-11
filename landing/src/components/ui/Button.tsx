import type { ReactNode } from 'react';
import { cn } from '../../lib/cn';

export type ButtonVariant =
  | 'primary'
  | 'secondary'
  | 'dark'
  | 'light'
  | 'outlineLight'
  | 'ghost';

type ButtonSize = 'sm' | 'md' | 'lg';

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    'bg-brand-yellow text-brand-espresso shadow-[0_14px_30px_-16px_rgba(244,180,0,0.95)] hover:bg-brand-yellow-dark',
  secondary:
    'border border-brand-espresso/15 bg-brand-white text-brand-espresso hover:border-brand-espresso/35 hover:bg-brand-cream',
  dark: 'bg-brand-espresso text-brand-cream shadow-[0_14px_30px_-18px_rgba(56,32,15,0.9)] hover:bg-[#2b180b]',
  light: 'bg-brand-cream text-brand-espresso hover:bg-brand-white',
  outlineLight:
    'border border-brand-cream/25 bg-transparent text-brand-cream hover:border-brand-cream/45 hover:bg-brand-cream/10',
  ghost:
    'bg-transparent text-brand-espresso hover:bg-brand-espresso/5',
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'gap-1.5 px-4 py-2 text-[0.82rem]',
  md: 'gap-2 px-5 py-3 text-[0.92rem]',
  lg: 'gap-2.5 px-6 py-3.5 text-[0.98rem]',
};

interface ButtonProps {
  children: ReactNode;
  href?: string;
  onClick?: () => void;
  type?: 'button' | 'submit';
  disabled?: boolean;
  external?: boolean;
  variant?: ButtonVariant;
  size?: ButtonSize;
  className?: string;
  icon?: ReactNode;
  iconRight?: ReactNode;
  ariaLabel?: string;
}

export function Button({
  children,
  href,
  onClick,
  type = 'button',
  disabled = false,
  external = false,
  variant = 'primary',
  size = 'md',
  className,
  icon,
  iconRight,
  ariaLabel,
}: ButtonProps) {
  const classes = cn(
    'group inline-flex select-none items-center justify-center rounded-full font-semibold tracking-[-0.01em] transition-all duration-200 ease-out hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-espresso focus-visible:ring-offset-2 focus-visible:ring-offset-brand-cream disabled:pointer-events-none disabled:opacity-60',
    variantClasses[variant],
    sizeClasses[size],
    className,
  );

  const content = (
    <>
      {icon ? (
        <span className="shrink-0" aria-hidden="true">
          {icon}
        </span>
      ) : null}
      <span>{children}</span>
      {iconRight ? (
        <span
          className="shrink-0 transition-transform duration-200 group-hover:translate-x-0.5"
          aria-hidden="true"
        >
          {iconRight}
        </span>
      ) : null}
    </>
  );

  if (href) {
    return (
      <a
        href={href}
        onClick={onClick}
        className={classes}
        aria-label={ariaLabel}
        {...(external
          ? { target: '_blank', rel: 'noopener noreferrer' }
          : {})}
      >
        {content}
      </a>
    );
  }

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={classes}
      aria-label={ariaLabel}
    >
      {content}
    </button>
  );
}
