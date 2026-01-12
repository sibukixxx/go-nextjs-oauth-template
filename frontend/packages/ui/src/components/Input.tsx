import { cn } from '../utils/cn';
import type { InputHTMLAttributes, ReactNode } from 'react';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  hint?: string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
}

export function Input({
  label,
  error,
  hint,
  leftIcon,
  rightIcon,
  className,
  id,
  ...props
}: InputProps) {
  const inputId = id || `input-${Math.random().toString(36).slice(2, 9)}`;

  return (
    <div className="w-full">
      {label && (
        <label
          htmlFor={inputId}
          className="block text-sm font-medium text-[var(--color-text-primary)] mb-1.5"
        >
          {label}
        </label>
      )}
      <div className="relative">
        {leftIcon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]">
            {leftIcon}
          </div>
        )}
        <input
          id={inputId}
          className={cn(
            'w-full px-4 py-2.5',
            'bg-[var(--color-bg-secondary)]',
            'border border-[var(--color-border)]',
            'rounded-[var(--radius-lg)]',
            'text-[var(--color-text-primary)]',
            'placeholder:text-[var(--color-text-tertiary)]',
            'transition-all duration-200',
            'focus:outline-none focus:ring-2 focus:ring-[var(--color-accent-light)] focus:border-[var(--color-accent)]',
            'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-[var(--color-bg-tertiary)]',
            error && 'border-[var(--color-error)] focus:ring-red-200 focus:border-[var(--color-error)]',
            leftIcon && 'pl-10',
            rightIcon && 'pr-10',
            className
          )}
          {...props}
        />
        {rightIcon && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]">
            {rightIcon}
          </div>
        )}
      </div>
      {error && (
        <p className="mt-1.5 text-sm text-[var(--color-error)]">{error}</p>
      )}
      {hint && !error && (
        <p className="mt-1.5 text-sm text-[var(--color-text-tertiary)]">{hint}</p>
      )}
    </div>
  );
}
