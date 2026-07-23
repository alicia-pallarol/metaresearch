/**
 * Client for the feedback API.
 *
 * Everything here is optional to the page: the curated map renders from the
 * bundle, and every function below fails soft. Free backends sleep, so a first
 * request can take the better part of a minute to wake the instance, that is a
 * normal state to be shown, not an error.
 */
import type { Summary } from '@/types/atlas'

const rawBase = (import.meta.env.VITE_API_BASE ?? '').trim()
export const apiBase = rawBase.replace(/\/+$/, '')
export const apiConfigured = apiBase.length > 0

export type ApiState = 'idle' | 'loading' | 'waking' | 'ready' | 'unavailable' | 'disabled'

export interface FeedbackPayload {
  /**
   * Where the form was opened: an agenda id, OR an area_tag + problem_id cell.
   * Exactly one shape. This is provenance, the votes below are what is claimed.
   */
  agenda_id?: string
  area_tag?: string
  problem_id?: string
  /** Required: the reader's maturity tier for the research area as a whole. */
  area_maturity: string
  /** Optional: agenda id -> maturity tier, for agendas of that same area. */
  subarea_maturity?: Record<string, string>
  familiarity: number
  notes?: string
  name?: string
  email?: string
  is_anonymous: boolean
  contact_consent: boolean
  reuse_consent: boolean
  submitter_token: string
  turnstile_token?: string
  /** Honeypot. Always sent empty by the real form. */
  website: string
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly field?: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init: RequestInit = {}, timeoutMs = 12000): Promise<T> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const res = await fetch(`${apiBase}${path}`, {
      ...init,
      signal: controller.signal,
      headers: { 'Content-Type': 'application/json', ...(init.headers ?? {}) },
    })
    const text = await res.text()
    const body: unknown = text ? JSON.parse(text) : {}

    if (!res.ok) {
      const detail = body as { error?: string; detail?: string; field?: string }
      throw new ApiError(detail.detail ?? detail.error ?? `Request failed (${res.status})`, res.status, detail.field)
    }
    return body as T
  } finally {
    clearTimeout(timer)
  }
}

/** Pings /api/health, which is also what wakes a cold-started instance. */
export async function health(timeoutMs = 60000): Promise<boolean> {
  if (!apiConfigured) return false
  try {
    await request<{ status: string }>('/api/health', { method: 'GET' }, timeoutMs)
    return true
  } catch {
    return false
  }
}

export async function fetchSummary(): Promise<Summary> {
  if (!apiConfigured) throw new ApiError('API is not configured', 0)
  return request<Summary>('/api/summary', { method: 'GET' }, 12000)
}

/**
 * Submits one piece of feedback. A cold instance refuses the first request, so a
 * network-level failure is retried once behind a health ping; onWaking lets the
 * UI say "waking the server…" instead of looking stuck.
 */
export async function submitFeedback(payload: FeedbackPayload, onWaking?: () => void): Promise<void> {
  if (!apiConfigured) throw new ApiError('Feedback is not available on this deployment', 0)

  const send = () => request<{ ok: boolean }>('/api/feedback', { method: 'POST', body: JSON.stringify(payload) }, 20000)

  try {
    await send()
  } catch (err) {
    // A validation or rate-limit answer is a real answer: do not retry it.
    if (err instanceof ApiError && err.status >= 400 && err.status !== 503) throw err

    onWaking?.()
    const awake = await health()
    if (!awake) {
      throw new ApiError('The feedback server is not responding. Your notes were not saved, please try again later.', 0)
    }
    await send()
  }
}
