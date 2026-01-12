import { cn } from '../utils/cn';
import { themes, themeLabels, themeDescriptions, applyTheme, type Theme } from '../themes';

export interface ThemeSelectorProps {
  currentTheme: Theme;
  onThemeChange?: (theme: Theme) => void;
  className?: string;
}

const themePreviewColors: Record<Theme, { bg: string; accent: string }> = {
  mint: { bg: '#E8F5F3', accent: '#2DD4BF' },
  cream: { bg: '#FDF8F3', accent: '#F59E0B' },
  lavender: { bg: '#F5F3FF', accent: '#A78BFA' },
};

export function ThemeSelector({
  currentTheme,
  onThemeChange,
  className,
}: ThemeSelectorProps) {
  const handleThemeChange = (theme: Theme) => {
    applyTheme(theme);
    onThemeChange?.(theme);
  };

  return (
    <div className={cn('grid grid-cols-1 sm:grid-cols-3 gap-4', className)}>
      {themes.map((theme) => {
        const isSelected = currentTheme === theme;
        const colors = themePreviewColors[theme];

        return (
          <button
            key={theme}
            onClick={() => handleThemeChange(theme)}
            className={cn(
              'p-4 rounded-[var(--radius-xl)] text-left transition-all duration-200',
              'border-2',
              isSelected
                ? 'border-[var(--color-accent)] shadow-[var(--shadow-md)]'
                : 'border-[var(--color-border)] hover:border-[var(--color-border-hover)]'
            )}
          >
            {/* Theme preview */}
            <div
              className="w-full h-16 rounded-[var(--radius-lg)] mb-3 flex items-end justify-end p-2"
              style={{ backgroundColor: colors.bg }}
            >
              <div
                className="w-8 h-8 rounded-[var(--radius-md)]"
                style={{ backgroundColor: colors.accent }}
              />
            </div>

            {/* Theme info */}
            <div className="flex items-center gap-2 mb-1">
              <span className="font-medium text-[var(--color-text-primary)]">
                {themeLabels[theme]}
              </span>
              {isSelected && (
                <svg
                  className="w-4 h-4 text-[var(--color-accent)]"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              )}
            </div>
            <p className="text-sm text-[var(--color-text-secondary)]">
              {themeDescriptions[theme]}
            </p>
          </button>
        );
      })}
    </div>
  );
}
