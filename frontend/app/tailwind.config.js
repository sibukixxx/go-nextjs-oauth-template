import uiPreset from '@app/ui/tailwind.preset';

/** @type {import('tailwindcss').Config} */
export default {
  presets: [uiPreset],
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
    '../packages/ui/src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
