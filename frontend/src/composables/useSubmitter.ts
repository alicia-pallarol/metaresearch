/**
 * The researcher's local identity.
 *
 * Everything here lives in localStorage on the reader's own machine: a random
 * token so repeat submissions can be counted as one person, plus whatever
 * name/email/consent they chose last time so later agendas only ask for
 * familiarity and notes. Nothing is sent anywhere until they submit a form.
 *
 * Clearing it is a one-click affordance in the UI, and clearing site data has
 * the same effect.
 */
import { computed, reactive, watch } from 'vue'

const STORAGE_KEY = 'atlas.submitter.v1'

export interface SubmitterProfile {
  token: string
  name: string
  email: string
  isAnonymous: boolean
  contactConsent: boolean
  reuseConsent: boolean
  /** Agenda ids this browser has already submitted for. */
  submitted: string[]
}

function newToken(): string {
  const c: Crypto | undefined = globalThis.crypto
  if (typeof c?.randomUUID === 'function') return c.randomUUID()

  // Fallback for browsers without randomUUID (older Safari, and any non-secure
  // context, where crypto.randomUUID is not exposed).
  const bytes = new Uint8Array(16)
  if (typeof c?.getRandomValues === 'function') c.getRandomValues(bytes)
  else for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256)

  bytes[6] = ((bytes[6] ?? 0) & 0x0f) | 0x40
  bytes[8] = ((bytes[8] ?? 0) & 0x3f) | 0x80
  const hex = [...bytes].map((b) => b.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

function emptyProfile(): SubmitterProfile {
  return {
    token: newToken(),
    name: '',
    email: '',
    isAnonymous: true,
    contactConsent: false,
    reuseConsent: false,
    submitted: [],
  }
}

function load(): SubmitterProfile {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return emptyProfile()
    const parsed = JSON.parse(raw) as Partial<SubmitterProfile>
    const base = emptyProfile()
    return {
      token: typeof parsed.token === 'string' && parsed.token.length === 36 ? parsed.token : base.token,
      name: typeof parsed.name === 'string' ? parsed.name : '',
      email: typeof parsed.email === 'string' ? parsed.email : '',
      isAnonymous: parsed.isAnonymous !== false,
      contactConsent: parsed.contactConsent === true,
      reuseConsent: parsed.reuseConsent === true,
      submitted: Array.isArray(parsed.submitted) ? parsed.submitted.filter((x): x is string => typeof x === 'string') : [],
    }
  } catch {
    // Private mode, disabled storage, corrupted value: start clean rather than break.
    return emptyProfile()
  }
}

const profile = reactive<SubmitterProfile>(load())

watch(
  () => ({ ...profile }),
  (value) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
    } catch {
      // Storage unavailable, the form still works, it just will not prefill.
    }
  },
  { deep: true },
)

export function useSubmitter() {
  const displayName = computed(() => {
    if (profile.isAnonymous) return 'anonymously'
    return profile.name.trim() || profile.email.trim() || 'without a name'
  })

  const hasIdentity = computed(() => !profile.isAnonymous && (profile.name.trim() !== '' || profile.email.trim() !== ''))

  function hasSubmitted(agendaId: string): boolean {
    return profile.submitted.includes(agendaId)
  }

  function markSubmitted(agendaId: string): void {
    if (!profile.submitted.includes(agendaId)) profile.submitted.push(agendaId)
  }

  function reset(): void {
    const fresh = emptyProfile()
    Object.assign(profile, fresh)
  }

  return { profile, displayName, hasIdentity, hasSubmitted, markSubmitted, reset }
}
