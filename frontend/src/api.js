// Thin API client with cold-start handling for the free Render tier, which
// sleeps after inactivity and can take 30–60s to wake.
import { API_BASE_URL } from './config.js'

class ApiError extends Error {
  constructor(message, { status = 0, fields = [] } = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.fields = fields
  }
}

async function request(path, { method = 'GET', body, signal } = {}) {
  let res
  try {
    res = await fetch(API_BASE_URL + path, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
      signal,
    })
  } catch (err) {
    if (err.name === 'AbortError') throw err
    throw new ApiError('network', { status: 0 })
  }

  let data = null
  const text = await res.text()
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = null
    }
  }

  if (!res.ok) {
    const msg = (data && data.error) || `Request failed (${res.status})`
    throw new ApiError(msg, { status: res.status, fields: (data && data.fields) || [] })
  }
  return data
}

// waitForWake polls /api/health until the backend responds or the attempts run
// out. Returns true once awake. Used to show a friendly cold-start state.
export async function waitForWake({ attempts = 20, intervalMs = 3000, onAttempt } = {}) {
  for (let i = 0; i < attempts; i++) {
    if (onAttempt) onAttempt(i + 1, attempts)
    try {
      const ctrl = new AbortController()
      const t = setTimeout(() => ctrl.abort(), 8000)
      const data = await request('/api/health', { signal: ctrl.signal })
      clearTimeout(t)
      if (data && data.status === 'ok') return true
    } catch {
      // keep waiting through cold-start timeouts
    }
    await new Promise((r) => setTimeout(r, intervalMs))
  }
  return false
}

export function getResearchAreas() {
  return request('/api/research-areas')
}

export function submitFamiliarity(payload) {
  return request('/api/submissions', { method: 'POST', body: payload })
}

export { ApiError }
