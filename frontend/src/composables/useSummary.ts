/**
 * Community feedback aggregates, shared by every component that needs them.
 *
 * One module-level store rather than a per-component fetch: the hero counter,
 * the familiarity overlay and each agenda's badge all read the same numbers.
 */
import { computed, reactive, readonly, ref } from 'vue'
import { apiConfigured, fetchSummary, type ApiState } from '@/lib/api'
import type { RatingSummary, Summary } from '@/types/atlas'

const summary = ref<Summary | null>(null)
const state = ref<ApiState>(apiConfigured ? 'idle' : 'disabled')
const lastError = ref<string | null>(null)

/** Optimistic local bumps, so a contributor sees their own submission at once. */
const pending = reactive<Record<string, number>>({})

let inFlight: Promise<void> | null = null

export async function refreshSummary(): Promise<void> {
  if (!apiConfigured) {
    state.value = 'disabled'
    return
  }
  if (inFlight) return inFlight

  state.value = summary.value ? 'ready' : 'loading'
  lastError.value = null

  inFlight = fetchSummary()
    .then((data) => {
      summary.value = data
      state.value = 'ready'
      for (const key of Object.keys(pending)) delete pending[key]
    })
    .catch((err: unknown) => {
      // Not an error the reader needs to act on: the map is fully usable without it.
      state.value = 'unavailable'
      lastError.value = err instanceof Error ? err.message : String(err)
    })
    .finally(() => {
      inFlight = null
    })

  return inFlight
}

export function noteLocalSubmission(subjectId: string): void {
  pending[subjectId] = (pending[subjectId] ?? 0) + 1
}

export function useSummary() {
  const totalSubmissions = computed(() => {
    const base = summary.value?.totals.total_submissions ?? 0
    return base + Object.values(pending).reduce((sum, n) => sum + n, 0)
  })

  const uniqueSubmitters = computed(() => summary.value?.totals.unique_submitters ?? 0)

  const hasData = computed(() => (summary.value?.totals.total_submissions ?? 0) > 0)

  /**
   * What the community voted for one research area: a maturity histogram and a
   * familiarity histogram. This is what the community view of the grid rests on.
   */
  function forArea(areaTag: string): RatingSummary | undefined {
    return summary.value?.per_area?.[areaTag]
  }

  /**
   * The same for one agenda, counting only the readers who rated that agenda
   * specifically. Undefined until somebody has.
   */
  function forAgenda(agendaId: string): RatingSummary | undefined {
    return summary.value?.per_agenda?.[agendaId]
  }

  return {
    // A computed rather than readonly(): the community grid reads the whole
    // structure, and a deep-readonly wrapper would only need casting away again.
    summary: computed(() => summary.value),
    state: readonly(state),
    lastError: readonly(lastError),
    totalSubmissions,
    uniqueSubmitters,
    hasData,
    forArea,
    forAgenda,
    refresh: refreshSummary,
  }
}
