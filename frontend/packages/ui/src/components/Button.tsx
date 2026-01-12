import { cn } from '../utils/cn';
import type { ButtonHTMLAttributes, ReactNode } from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  children: ReactNode;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary: [
    'bg-[var(--color-accent)] text-[var(--color-text-inverse)]',
    'hover:bg-[var(--color-accent-hover)]',
    'focus:ring-[var(--color-accent-light)]',
  ].join(' '),
  secondary: [
    'bg-[var(--color-bg-secondary)] text-[var(--color-text-primary)]',
    'border border-[var(--color-border)]',
    'hover:border-[var(--color-border-hover)] hover:bg-[var(--color-bg-tertiary)]',
    'focus:ring-[var(--color-border)]',
  ].join(' '),
  ghost: [
    'bg-transparent text-[var(--color-text-primary)]',
    'hover:bg-[var(--color-bg-tertiary)]',
    'focus:ring-[var(--color-border)]',
  ].join(' '),
  danger: [
    'bg-[var(--color-error)] text-white',
    'hover:bg-red-600',
    'focus:ring-red-200',
  ].join(' '),
};

const sizeStyles: Record<ButtonSize, string> = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-4 py-2 text-base',
  lg: 'px-6 py-3 text-lg',
};

export function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center gap-2',
        'font-medium rounded-[var(--radius-lg)]',
        'transition-all duration-200 ease-out',
        'focus:outline-none focus:ring-2 focus:ring-offset-2',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        variantStyles[variant],
        sizeStyles[size],
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg
          className="animate-spin h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
}
