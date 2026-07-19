// Shared, cached aggregates state. Both the framework/heatmap page and the
// Swiss cheese page read from the same source so they stay consistent.
import { reactive, readonly } from 'vue'
import { getResearchAreas, waitForWake } from '../api.js'

const state = reactive({
  loading: false,
  waking: false, // true while polling a cold-started backend
  wakeAttempt: 0,
  wakeTotal: 0,
  error: '',
  loaded: false,
  data: null, // { total_submissions, low_sample_threshold, areas: [...] }
})

async function load(force = false) {
  if (state.loading) return
  if (state.loaded && !force) return
  state.loading = true
  state.error = ''
  try {
    // First attempt: if it fails fast (likely a sleeping backend), show the
    // "waking the server" state and poll health until it responds.
    try {
      state.data = await getResearchAreas()
    } catch (first) {
      state.waking = true
      const awake = await waitForWake({
        onAttempt: (n, total) => {
          state.wakeAttempt = n
          state.wakeTotal = total
        },
      })
      state.waking = false
      if (!awake) throw first
      state.data = await getResearchAreas()
    }
    state.loaded = true
  } catch (err) {
    state.error =
      'We could not reach the server. It may be waking from sleep — please try again in a moment.'
  } finally {
    state.loading = false
  }
}

export function useAggregates() {
  return {
    state: readonly(state),
    load,
    reload: () => load(true),
  }
}
