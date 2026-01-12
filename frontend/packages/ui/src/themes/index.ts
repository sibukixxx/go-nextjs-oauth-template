export type Theme = 'mint' | 'cream' | 'lavender';

export const themes: Theme[] = ['mint', 'cream', 'lavender'];

export const themeLabels: Record<Theme, string> = {
  mint: 'Mint',
  cream: 'Cream',
  lavender: 'Lavender',
};

export const themeDescriptions: Record<Theme, string> = {
  mint: 'Fresh and calming light blue tones',
  cream: 'Warm and cozy beige tones',
  lavender: 'Soft and gentle purple tones',
};

const THEME_STORAGE_KEY = 'app-theme';

export function getStoredTheme(): Theme | null {
  if (typeof window === 'undefined') return null;
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored && themes.includes(stored as Theme)) {
    return stored as Theme;
  }
  return null;
}

export function setStoredTheme(theme: Theme): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(THEME_STORAGE_KEY, theme);
}

export function applyTheme(theme: Theme): void {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute('data-theme', theme);
  setStoredTheme(theme);
}

export function initializeTheme(): Theme {
  const stored = getStoredTheme();
  const theme = stored ?? 'mint';
  applyTheme(theme);
  return theme;
}
