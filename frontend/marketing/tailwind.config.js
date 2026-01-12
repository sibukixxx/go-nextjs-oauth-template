/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Noto Sans JP', 'Nunito', 'sans-serif'],
        display: ['Nunito', 'sans-serif'],
      },
      colors: {
        primary: 'var(--color-accent)',
        'primary-hover': 'var(--color-accent-hover)',
        'primary-light': 'var(--color-accent-light)',
        'primary-dark': 'var(--color-accent-dark)',
        background: 'var(--color-bg-primary)',
        'background-secondary': 'var(--color-bg-secondary)',
        'background-tertiary': 'var(--color-bg-tertiary)',
        foreground: 'var(--color-text-primary)',
        'foreground-secondary': 'var(--color-text-secondary)',
        'foreground-tertiary': 'var(--color-text-tertiary)',
        border: 'var(--color-border)',
        'border-hover': 'var(--color-border-hover)',
      },
    },
  },
  plugins: [],
};
