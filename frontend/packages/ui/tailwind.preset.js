/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['Noto Sans JP', 'Nunito', 'sans-serif'],
        display: ['Nunito', 'sans-serif'],
      },
      colors: {
        // Map CSS variables to Tailwind classes
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

        success: 'var(--color-success)',
        warning: 'var(--color-warning)',
        error: 'var(--color-error)',
        info: 'var(--color-info)',
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        xl: 'var(--radius-xl)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
      },
    },
  },
};
