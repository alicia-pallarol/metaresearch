// Runtime configuration. The API base URL is injected at build time via the
// VITE_API_BASE_URL environment variable so the same code can point at a
// local backend, a Render service, or a future custom domain with no changes.
//
// In local dev, leaving it unset defaults to the local Go server.
const raw = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// Normalize: strip any trailing slash so we can safely concatenate paths.
export const API_BASE_URL = raw.replace(/\/+$/, '')

// Familiarity scale, shown as the legend across the app.
export const FAMILIARITY_LEVELS = [
  { value: 0, label: 'No familiarity', short: 'None' },
  { value: 1, label: 'Basic (have heard of the area)', short: 'Basic' },
  { value: 2, label: 'Medium (understand the concepts, have read about it)', short: 'Medium' },
  { value: 3, label: 'High (could actively work or research in this area)', short: 'High' },
]
