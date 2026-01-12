import { cn } from '../utils/cn';
import type { HTMLAttributes, ReactNode } from 'react';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
}

const paddingStyles = {
  none: '',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
};

export function Card({
  children,
  padding = 'md',
  hover = false,
  className,
  ...props
}: CardProps) {
  return (
    <div
      className={cn(
        'bg-[var(--color-bg-secondary)]',
        'border border-[var(--color-border)]',
        'rounded-[var(--radius-xl)]',
        'shadow-[var(--shadow-sm)]',
        hover && 'transition-shadow duration-200 hover:shadow-[var(--shadow-md)]',
        paddingStyles[padding],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function CardHeader({ children, className, ...props }: CardHeaderProps) {
  return (
    <div
      className={cn('pb-4 border-b border-[var(--color-border)]', className)}
      {...props}
    >
      {children}
    </div>
  );
}

export interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement> {
  children: ReactNode;
}

export function CardTitle({ children, className, ...props }: CardTitleProps) {
  return (
    <h3
      className={cn(
        'text-lg font-semibold text-[var(--color-text-primary)]',
        className
      )}
      {...props}
    >
      {children}
    </h3>
  );
}

export interface CardDescriptionProps extends HTMLAttributes<HTMLParagraphElement> {
  children: ReactNode;
}

export function CardDescription({ children, className, ...props }: CardDescriptionProps) {
  return (
    <p
      className={cn('text-sm text-[var(--color-text-secondary)] mt-1', className)}
      {...props}
    >
      {children}
    </p>
  );
}

export interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function CardContent({ children, className, ...props }: CardContentProps) {
  return (
    <div className={cn('pt-4', className)} {...props}>
      {children}
    </div>
  );
}

export interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function CardFooter({ children, className, ...props }: CardFooterProps) {
  return (
    <div
      className={cn(
        'pt-4 mt-4 border-t border-[var(--color-border)] flex items-center gap-3',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
