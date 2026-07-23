/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>
  export default component
}

interface ImportMetaEnv {
  /** Base URL of the Go API, e.g. https://ai-safety-atlas-api.onrender.com. Empty disables the API. */
  readonly VITE_API_BASE?: string
  /** Optional Cloudflare Turnstile site key. When absent no widget is rendered. */
  readonly VITE_TURNSTILE_SITE_KEY?: string
  /** Contact address shown in the privacy notice (GDPR controller contact). */
  readonly VITE_CONTACT_EMAIL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
