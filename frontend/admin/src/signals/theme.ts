import { signal } from '@preact/signals';

export type Theme = 'mint' | 'cream' | 'lavender';

const THEME_STORAGE_KEY = 'app-theme';

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'mint';
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  if (stored && ['mint', 'cream', 'lavender'].includes(stored)) {
    return stored as Theme;
  }
  return 'mint';
}

export const currentTheme = signal<Theme>(getStoredTheme());

export function setTheme(theme: Theme) {
  currentTheme.value = theme;
  localStorage.setItem(THEME_STORAGE_KEY, theme);
  document.documentElement.setAttribute('data-theme', theme);
}

export function initializeTheme() {
  const theme = getStoredTheme();
  document.documentElement.setAttribute('data-theme', theme);
  currentTheme.value = theme;
}

export const themes: { value: Theme; label: string; description: string }[] = [
  { value: 'mint', label: 'Mint', description: 'Fresh and calming light blue tones' },
  { value: 'cream', label: 'Cream', description: 'Warm and cozy beige tones' },
  { value: 'lavender', label: 'Lavender', description: 'Soft and gentle purple tones' },
];
