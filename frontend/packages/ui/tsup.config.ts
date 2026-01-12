import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts', 'src/themes/index.ts'],
  format: ['esm'],
  dts: true,
  clean: true,
  external: ['react', 'react-dom'],
  esbuildOptions(options) {
    options.jsx = 'automatic';
  },
});
