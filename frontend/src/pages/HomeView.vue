<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAggregates } from '../composables/useAggregates.js'
import FamiliarityLegend from '../components/FamiliarityLegend.vue'
import FrameworkTable from '../components/FrameworkTable.vue'
import HeatmapGrid from '../components/HeatmapGrid.vue'
import LoadingState from '../components/LoadingState.vue'
import SubmissionForm from '../components/SubmissionForm.vue'

const { state, load, reload } = useAggregates()

const view = ref('table') // 'table' | 'heatmap'
const showForm = ref(false)
const flash = ref('')

const areas = computed(() => state.data?.areas || [])
const total = computed(() => state.data?.total_submissions || 0)
const lowSample = computed(() => state.data?.low_sample_threshold ?? 3)
const hasAnyRatings = computed(() => areas.value.some((a) => a.n > 0))

onMounted(load)

function onSubmitted({ replaced }) {
  showForm.value = false
  flash.value = replaced
    ? 'Your previous response was updated. Thank you.'
    : 'Your response was recorded. Thank you.'
  reload()
  setTimeout(() => (flash.value = ''), 6000)
}
</script>

<template>
  <div class="container page">
    <section class="intro">
      <h1>A community map of AI safety research familiarity</h1>
      <p class="lede">
        Researchers self-report their familiarity with each area of a framework of
        AI safety research. Responses are aggregated into the heatmap below to show
        where collective expertise is concentrated — and where it is thin.
      </p>
      <div class="intro-actions">
        <button class="btn btn-primary" @click="showForm = true">Submit your familiarity</button>
        <RouterLink class="btn" to="/swiss-cheese">View the defense model</RouterLink>
      </div>
    </section>

    <div v-if="flash" class="notice flash" role="status">{{ flash }}</div>

    <section class="controls card" aria-label="Legend and view controls">
      <FamiliarityLegend :show-scale="true" />
      <div class="control-right">
        <div class="sample" :class="{ low: total > 0 && total < lowSample }">
          <strong>n = {{ total }}</strong>
          <span class="small muted">{{ total === 1 ? 'researcher' : 'researchers' }}</span>
        </div>
        <div class="viewtoggle" role="group" aria-label="View">
          <button :class="{ active: view === 'table' }" :aria-pressed="view === 'table'" @click="view = 'table'">
            Table
          </button>
          <button :class="{ active: view === 'heatmap' }" :aria-pressed="view === 'heatmap'" @click="view = 'heatmap'">
            Heatmap
          </button>
        </div>
      </div>
    </section>

    <p v-if="total > 0 && total < lowSample" class="small low-note">
      Sample size is small (n &lt; {{ lowSample }}). Values are shown but should be
      read as provisional.
    </p>

    <!-- Loading / cold-start -->
    <LoadingState
      v-if="state.loading && !state.loaded"
      :waking="state.waking"
      :attempt="state.wakeAttempt"
      :total="state.wakeTotal"
    />

    <!-- Error -->
    <div v-else-if="state.error && !state.loaded" class="notice notice-warn error-box" role="alert">
      <p>{{ state.error }}</p>
      <button class="btn" @click="reload">Try again</button>
    </div>

    <!-- Empty state: framework loaded but no ratings yet -->
    <template v-else>
      <div v-if="!hasAnyRatings" class="empty card">
        <h2>No submissions yet</h2>
        <p class="muted">
          The framework of research areas is shown below. Be the first to record
          your familiarity — the heatmap fills in as researchers submit.
        </p>
      </div>

      <FrameworkTable v-if="view === 'table'" :areas="areas" :low-sample="lowSample" />
      <HeatmapGrid v-else :areas="areas" :low-sample="lowSample" />
    </template>

    <SubmissionForm
      v-if="showForm"
      :areas="areas"
      @close="showForm = false"
      @submitted="onSubmitted"
    />
  </div>
</template>

<style scoped>
.page {
  padding-top: 32px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.intro {
  max-width: 74ch;
}
.intro h1 {
  font-size: clamp(1.6rem, 3.5vw, 2.3rem);
  margin: 0 0 10px;
}
.lede {
  font-size: 1.08rem;
  color: var(--text-secondary);
  margin: 0 0 18px;
}
.intro-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.flash {
  border-left-color: var(--good);
  color: var(--text-primary);
}
.controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
}
.control-right {
  display: flex;
  align-items: center;
  gap: 18px;
}
.sample {
  display: flex;
  align-items: baseline;
  gap: 6px;
}
.sample.low strong {
  color: #9a6a00;
}
.viewtoggle {
  display: inline-flex;
  border: 1px solid var(--border-strong);
  border-radius: 7px;
  overflow: hidden;
}
.viewtoggle button {
  padding: 7px 14px;
  border: 0;
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  font-weight: 500;
}
.viewtoggle button.active {
  background: var(--accent);
  color: var(--accent-ink);
}
.low-note {
  color: #9a6a00;
  margin: -4px 2px 0;
}
@media (prefers-color-scheme: dark) {
  .low-note {
    color: #e0b050;
  }
  .sample.low strong {
    color: #e0b050;
  }
}
.error-box {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}
.empty {
  padding: 22px;
}
.empty h2 {
  margin: 0 0 6px;
  font-size: 1.2rem;
}
.empty p {
  margin: 0;
  max-width: 60ch;
}
</style>
