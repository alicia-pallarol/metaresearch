import { fileURLToPath, URL } from 'node:url'
import { defineConfig, mergeConfig } from 'vitest/config'
import viteConfig from './vite.config'

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      environment: 'jsdom',
      include: ['src/**/*.test.ts'],
      root: fileURLToPath(new URL('./', import.meta.url)),
      // Tests run as if an API were configured; the "no API configured" and
      // "API unreachable" paths are exercised explicitly where they matter.
      env: { VITE_API_BASE: 'http://api.test' },
    },
  }),
)
