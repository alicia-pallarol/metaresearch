import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The curated dataset lives at the repo root, outside the frontend package, so
// there is exactly one copy of it. It is aliased in (and bundled) rather than
// fetched, which is what lets the map render with the backend unreachable.
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@data': fileURLToPath(new URL('../data', import.meta.url)),
      '@root': fileURLToPath(new URL('..', import.meta.url)),
    },
  },
  server: {
    fs: {
      // Allow the dev server to read ../data and ../swiss_layers.json.
      allow: [fileURLToPath(new URL('..', import.meta.url))],
    },
  },
  build: {
    target: 'es2020',
    sourcemap: false,
  },
})
