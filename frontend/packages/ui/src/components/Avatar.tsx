import { cn } from '../utils/cn';
import type { ImgHTMLAttributes } from 'react';

export type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

export interface AvatarProps extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> {
  src?: string | null;
  name?: string;
  size?: AvatarSize;
}

const sizeStyles: Record<AvatarSize, string> = {
  xs: 'w-6 h-6 text-xs',
  sm: 'w-8 h-8 text-sm',
  md: 'w-10 h-10 text-base',
  lg: 'w-12 h-12 text-lg',
  xl: 'w-16 h-16 text-xl',
};

function getInitials(name?: string): string {
  if (!name) return '?';
  const words = name.trim().split(' ');
  if (words.length === 1) {
    return words[0].charAt(0).toUpperCase();
  }
  return (words[0].charAt(0) + words[words.length - 1].charAt(0)).toUpperCase();
}

function getColorFromName(name?: string): string {
  if (!name) return 'bg-[var(--color-accent)]';
  const colors = [
    'bg-teal-500',
    'bg-amber-500',
    'bg-violet-500',
    'bg-rose-500',
    'bg-emerald-500',
    'bg-blue-500',
    'bg-orange-500',
    'bg-indigo-500',
  ];
  const index = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  return colors[index % colors.length];
}

export function Avatar({ src, name, size = 'md', className, alt, ...props }: AvatarProps) {
  const initials = getInitials(name);
  const bgColor = getColorFromName(name);

  if (src) {
    return (
      <img
        src={src}
        alt={alt || name || 'Avatar'}
        className={cn(
          'rounded-full object-cover',
          'ring-2 ring-[var(--color-bg-secondary)]',
          sizeStyles[size],
          className
        )}
        {...props}
      />
    );
  }

  return (
    <div
      className={cn(
        'rounded-full flex items-center justify-center',
        'font-medium text-white',
        'ring-2 ring-[var(--color-bg-secondary)]',
        bgColor,
        sizeStyles[size],
        className
      )}
      title={name}
    >
      {initials}
    </div>
  );
}

export interface AvatarGroupProps {
  children: React.ReactNode;
  max?: number;
  size?: AvatarSize;
  className?: string;
}

export function AvatarGroup({ children, max, size = 'md', className }: AvatarGroupProps) {
  const childArray = Array.isArray(children) ? children : [children];
  const visibleChildren = max ? childArray.slice(0, max) : childArray;
  const remaining = max ? childArray.length - max : 0;

  return (
    <div className={cn('flex -space-x-2', className)}>
      {visibleChildren}
      {remaining > 0 && (
        <div
          className={cn(
            'rounded-full flex items-center justify-center',
            'font-medium text-[var(--color-text-primary)]',
            'bg-[var(--color-bg-tertiary)]',
            'ring-2 ring-[var(--color-bg-secondary)]',
            sizeStyles[size]
          )}
        >
          +{remaining}
        </div>
      )}
    </div>
  );
}
